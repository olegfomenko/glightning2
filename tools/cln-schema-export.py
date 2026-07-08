#!/usr/bin/env python3
"""Dump Core Lightning msggen's normalized schema model as JSON.

This script deliberately does not make Go-specific decisions. It only loads
Core Lightning schemas through msggen, applies msggen's own patches, and
serializes the resulting msggen model objects into JSON-shaped dictionaries.
"""

from __future__ import annotations

import argparse
import json
import os
import sys
from pathlib import Path
from typing import Any

from msggen.model import (
    ArrayField,
    CompositeField,
    EnumField,
    PrimitiveField,
    UnionField,
)
from msggen.patch import OptionalPatch, OverridePatch, VersionAnnotationPatch
from msggen.utils import load_jsonrpc_service


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


def dump_service(service: Any) -> dict[str, Any]:
    return {
        "name": service.name,
        "methods": [dump_method(method) for method in service.methods],
        "notifications": [
            dump_notification(notification)
            for notification in service.notifications
        ],
        "hooks": [dump_hook(hook) for hook in service.hooks],
        "includes": list(service.includes),
    }


def dump_method(method: Any) -> dict[str, Any]:
    return {
        "name": method.name,
        "name_raw": method.name_raw,
        "request": dump_field(method.request),
        "response": dump_field(method.response),
    }


def dump_notification(notification: Any) -> dict[str, Any]:
    return {
        "name": notification.name,
        "typename": str(notification.typename),
        "request": dump_field(notification.request),
        "response": dump_field(notification.response),
    }


def dump_hook(hook: Any) -> dict[str, Any]:
    return {
        "name": hook.name,
        "typename": str(hook.typename),
        "request": dump_field(hook.request),
        "response": dump_field(hook.response),
    }


def dump_field(field: Any) -> dict[str, Any]:
    data = dump_common_field(field)

    if isinstance(field, PrimitiveField):
        data.update(
            {
                "kind": "primitive",
                "typename": field.typename,
            }
        )
    elif isinstance(field, EnumField):
        data.update(
            {
                "kind": "enum",
                "typename": str(field.typename),
                "variants": [str(variant) for variant in field.variants],
            }
        )
    elif isinstance(field, CompositeField):
        data.update(
            {
                "kind": "composite",
                "typename": str(field.typename),
                "fields": [dump_field(child) for child in field.fields],
            }
        )
    elif isinstance(field, ArrayField):
        data.update(
            {
                "kind": "array",
                "dims": field.dims,
                "itemtype": dump_field(field.itemtype),
            }
        )
    elif isinstance(field, UnionField):
        data.update(
            {
                "kind": "union",
                "typename": str(field.typename),
                "variants": [
                    dump_field(variant)
                    for variant in field.variants
                ],
            }
        )
    else:
        data.update({"kind": type(field).__name__})

    return data


def dump_common_field(field: Any) -> dict[str, Any]:
    return {
        "path": field.path,
        "name": str(field.name),
        "normalized": field.normalized(),
        "description": field.description,
        "required": bool(getattr(field, "required", False)),
        "optional": bool(getattr(field, "optional", False)),
        "added": getattr(field, "added", None),
        "deprecated": getattr(field, "deprecated", None),
        "omitted": bool(field.omit()),
        "override": field.override(),
    }


def main() -> int:
    args = parse_args()
    cln_repo = args.cln_repo.resolve()
    meta_path = cln_repo / ".msggen.json"
    if not meta_path.exists():
        print(f"error: metadata file not found: {meta_path}", file=sys.stderr)
        return 2

    old_cwd = Path.cwd()
    try:
        os.chdir(cln_repo)
        service = load_service(meta_path)
        payload = dump_service(service)
    finally:
        os.chdir(old_cwd)

    output = json.dumps(payload, indent=2, sort_keys=True)
    args.out.parent.mkdir(parents=True, exist_ok=True)
    args.out.write_text(output + "\n")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
