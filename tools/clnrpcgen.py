#!/usr/bin/env python3
"""Generate Go structs from Core Lightning msggen schemas."""

from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from pathlib import Path
from typing import Any

from msggen.gen.generator import IGenerator
from msggen.model import ArrayField, CompositeField, EnumField, PrimitiveField, UnionField
from msggen.patch import OptionalPatch, OverridePatch, VersionAnnotationPatch
from msggen.utils import load_jsonrpc_service


INITIALISMS = {
    "api",
    "bip",
    "cln",
    "csv",
    "db",
    "hex",
    "htlc",
    "id",
    "json",
    "msat",
    "psbt",
    "rpc",
    "scid",
    "tlv",
    "tx",
    "txid",
    "uri",
    "url",
    "utxo",
}


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("cln_repo", type=Path)
    parser.add_argument("out", type=Path)
    return parser.parse_args()


def load_service(meta_path: Path) -> Any:
    meta = json.loads(meta_path.read_text())
    service = load_jsonrpc_service()

    VersionAnnotationPatch(meta=meta).apply(service)
    OptionalPatch().apply(service)
    OverridePatch().apply(service)

    return service


class GoGenerator(IGenerator):
    def __init__(self, out_dir: Path, module_path: str, cln_commit: str) -> None:
        self.out_dir = out_dir
        self.module_path = module_path
        self.cln_commit = cln_commit
        self.defs: dict[str, str] = {}
        self.order: list[str] = []
        self.imports: set[str] = set()

    def generate(self, service: Any) -> None:
        for method in service.methods:
            self.collect_field(method.request)
            self.collect_field(method.response)
        for notification in service.notifications:
            self.collect_field(notification.request)
            self.collect_field(notification.response)
        for hook in service.hooks:
            self.collect_field(hook.request)
            self.collect_field(hook.response)

        self.out_dir.mkdir(parents=True, exist_ok=True)
        structs = self.out_dir / "clnrpc_generated.go"
        calls = self.out_dir / "calls_generated.go"
        structs.write_text(self.render_structs())
        calls.write_text(self.render_calls(service))
        subprocess.run(["gofmt", "-w", str(structs), str(calls)], check=True)

    def collect_field(self, field: Any) -> None:
        if field.omit():
            return

        if isinstance(field, CompositeField):
            for child in field.fields:
                self.collect_field(child)
            self.add_struct(field)
        elif isinstance(field, EnumField):
            self.add_enum(field)
        elif isinstance(field, ArrayField):
            self.collect_field(field.itemtype)
        elif isinstance(field, UnionField):
            for variant in field.variants:
                self.collect_field(variant)

    def add_struct(self, field: CompositeField) -> None:
        name = go_named_type(field)
        if not name or name in self.defs:
            return

        lines = []
        lines.extend(comment_lines(name, field.description))
        lines.append(f"type {name} struct {{")
        for child in field.fields:
            if child.omit():
                continue
            field_name = go_field_name(child.normalized())
            if not field_name:
                continue
            field_type = self.go_field_type(child)
            tag = json_tag(child)
            lines.extend(comment_lines(field_name, child.description, indent="\t"))
            lines.append(f"\t{field_name} {field_type} `{tag}`")
        lines.append("}")
        lines.append("")

        self.defs[name] = "\n".join(lines) + "\n"
        self.order.append(name)

    def add_enum(self, field: EnumField) -> None:
        name = go_named_type(field)
        if not name or name in self.defs:
            return

        lines = []
        lines.extend(comment_lines(name, field.description))
        lines.append(f"type {name} string")
        lines.append("")
        if field.variants:
            lines.append("const (")
            used: set[str] = set()
            for variant in field.variants:
                value = str(variant)
                const_name = unique_name(name + go_exported_name(value), used)
                lines.append(f"\t{const_name} {name} = {json.dumps(value)}")
            lines.append(")")
            lines.append("")

        self.defs[name] = "\n".join(lines) + "\n"
        self.order.append(name)

    def go_type(self, field: Any) -> str:
        if field.omit():
            return "any"

        if isinstance(field, PrimitiveField):
            return self.primitive_type(field.typename)
        if isinstance(field, (EnumField, CompositeField)):
            return go_named_type(field)
        if isinstance(field, ArrayField):
            dims = field.dims if field.dims > 0 else 1
            return "[]" * dims + self.go_type(field.itemtype)
        if isinstance(field, UnionField):
            return "any"
        return "any"

    def go_field_type(self, field: Any) -> str:
        field_type = self.go_type(field)
        if getattr(field, "optional", False):
            return f"*{field_type}"
        return field_type

    def primitive_type(self, name: str) -> str:
        builtin = {
            "boolean": "bool",
            "u8": "uint8",
            "u16": "uint16",
            "u32": "uint32",
            "u64": "uint64",
            "integer": "int64",
            "number": "float64",
            "float": "float64",
            "f32": "float32",
            "string": "string",
            "string_map": "map[string]string",
            "json_object_or_array": "any",
            "json_scalar": "any",
        }
        if name in builtin:
            return builtin[name]

        clntypes = {
            "hex": "Hex",
            "hash": "Hash",
            "secret": "Secret",
            "pubkey": "PubKey",
            "txid": "TxID",
            "outpoint": "Outpoint",
            "short_channel_id": "ShortChannelID",
            "short_channel_id_dir": "ShortChannelIDDir",
            "signature": "Signature",
            "bip340sig": "Bip340Sig",
            "currency": "Currency",
            "feerate": "Feerate",
            "outputdesc": "OutputDesc",
            "msat": "MSat",
            "sat": "Sat",
            "msat_or_all": "MSatOrAll",
            "msat_or_any": "AmountOrAny",
            "sat_or_all": "AmountOrAll",
        }
        if name in clntypes:
            self.imports.add(f"{self.module_path}/clntypes")
            return f"clntypes.{clntypes[name]}"

        return "any"

    def render_structs(self) -> str:
        lines = [
            f"// Code generated by tools/clnrpcgen.py from Core Lightning {self.cln_commit}; DO NOT EDIT.",
            "",
            "package clnrpc",
            "",
        ]

        if self.imports:
            lines.append("import (")
            for path in sorted(self.imports):
                lines.append(f"\t{json.dumps(path)}")
            lines.append(")")
            lines.append("")

        for name in self.order:
            lines.append(self.defs[name].rstrip())
            lines.append("")

        return "\n".join(lines).rstrip() + "\n"

    def render_calls(self, service: Any) -> str:
        lines = [
            f"// Code generated by tools/clnrpcgen.py from Core Lightning {self.cln_commit}; DO NOT EDIT.",
            "",
            "package clnrpc",
            "",
            'import "context"',
            "",
        ]

        for method in service.methods:
            req = go_named_type(method.request)
            res = go_named_type(method.response)
            name = req.removesuffix("Request")
            rpc_method = method.name_raw.lower()
            has_fields = any(not field.omit() for field in method.request.fields)

            lines.append(f"// {name} calls the {rpc_method} Core Lightning RPC method.")
            if has_fields:
                lines.append(
                    f"func (c *Client) {name}(ctx context.Context, req {req}) (*{res}, error) {{"
                )
                params = "req"
            else:
                lines.append(
                    f"func (c *Client) {name}(ctx context.Context) (*{res}, error) {{"
                )
                params = f"{req}{{}}"
            lines.append(f"\tvar result {res}")
            lines.append(
                f'\tif err := c.Call(ctx, "{rpc_method}", {params}, &result); err != nil {{'
            )
            lines.append("\t\treturn nil, err")
            lines.append("\t}")
            lines.append("\treturn &result, nil")
            lines.append("}")
            lines.append("")

        return "\n".join(lines).rstrip() + "\n"


def go_named_type(field: Any) -> str:
    name = field.path.removesuffix("[]")
    typename = str(getattr(field, "typename", ""))
    if typename.endswith("Request"):
        return go_exported_name(name) + "Request"
    if typename.endswith("Response"):
        return go_exported_name(name) + "Response"
    return go_exported_name(name)


def go_field_name(name: str) -> str:
    return go_exported_name(name.removesuffix("[]"))


def go_exported_name(name: str) -> str:
    parts = split_name(normalize_known_name(name))
    out = "".join(format_go_word(part) for part in parts)
    if out and out[0].isdigit():
        return "N" + out
    return out


def format_go_word(word: str) -> str:
    lower = word.lower()
    if lower in INITIALISMS:
        return lower.upper()
    return lower[:1].upper() + lower[1:].lower()


def split_name(name: str) -> list[str]:
    name = name.replace("[]", "")
    raw_parts = re.split(r"[^A-Za-z0-9]+", name)
    parts: list[str] = []
    for raw in raw_parts:
        parts.extend(split_camel(raw))
    return [part for part in parts if part]


def normalize_known_name(name: str) -> str:
    return name.replace("Getinfo", "GetInfo").replace("getinfo", "get_info")


def split_camel(name: str) -> list[str]:
    if not name:
        return []

    parts: list[str] = []
    start = 0
    for i in range(1, len(name)):
        prev = name[i - 1]
        cur = name[i]
        next_lower = i + 1 < len(name) and name[i + 1].islower()
        if (prev.islower() or prev.isdigit()) and cur.isupper():
            parts.append(name[start:i])
            start = i
        elif prev.isupper() and cur.isupper() and next_lower:
            parts.append(name[start:i])
            start = i
    parts.append(name[start:])
    return parts


def json_tag(field: Any) -> str:
    name = str(field.name).removesuffix("[]")
    if not name:
        name = field.path.split(".")[-1].removesuffix("[]")
    suffix = ",omitempty" if getattr(field, "optional", False) else ""
    return f'json:"{name}{suffix}"'


def comment_lines(name: str, value: Any, indent: str = "") -> list[str]:
    text = description(value)
    if not text:
        return []
    return [f"{indent}// {name} {text}"]


def description(value: Any) -> str:
    if value is None:
        return ""
    if isinstance(value, list):
        value = " ".join(str(item) for item in value)
    return " ".join(str(value).split())


def unique_name(name: str, used: set[str]) -> str:
    name = name or "Value"
    base = name
    i = 2
    while name in used:
        name = f"{base}{i}"
        i += 1
    used.add(name)
    return name


def module_path(repo_root: Path) -> str:
    for line in (repo_root / "go.mod").read_text().splitlines():
        if line.startswith("module "):
            return line.split()[1]
    raise RuntimeError("go.mod does not contain module path")


def git_commit(repo: Path) -> str:
    return subprocess.check_output(
        ["git", "-C", str(repo), "rev-parse", "HEAD"], text=True
    ).strip()


def main() -> int:
    args = parse_args()
    cln_repo = args.cln_repo.resolve()
    meta_path = cln_repo / ".msggen.json"
    if not meta_path.exists():
        print(f"error: metadata file not found: {meta_path}", file=sys.stderr)
        return 2

    repo_root = Path(__file__).resolve().parent.parent
    generator = GoGenerator(args.out.resolve(), module_path(repo_root), git_commit(cln_repo))

    old_cwd = Path.cwd()
    try:
        os.chdir(cln_repo)
        service = load_service(meta_path)
        generator.generate(service)
    finally:
        os.chdir(old_cwd)

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
