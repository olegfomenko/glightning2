#!/usr/bin/env bash
set -euo pipefail

cln_repo="$1"
out_dir="$2"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "${script_dir}/.." && pwd)"
tmp_dir="$(mktemp -d)"

trap 'rm -rf "${tmp_dir}"' EXIT

schema_json="${tmp_dir}/cln-msggen.json"
go_cache="${tmp_dir}/gocache"
go_tmp="${tmp_dir}/gotmp"

mkdir -p "${go_cache}" "${go_tmp}"

PYTHONPATH="${cln_repo}/contrib/msggen" \
  python3 "${script_dir}/cln-schema-export.py" "${cln_repo}" "${schema_json}"

(cd "${repo_root}" && GOTOOLCHAIN=local GOCACHE="${go_cache}" GOTMPDIR="${go_tmp}" go run -mod=vendor ./cmd/clnrpcgen "${schema_json}" "${out_dir}")
