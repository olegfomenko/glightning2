#!/usr/bin/env bash
set -euo pipefail

cln_repo="$1"
out_dir="$2"

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

PYTHONPATH="${cln_repo}/contrib/msggen" \
  python3 "${script_dir}/clnrpcgen.py" "${cln_repo}" "${out_dir}"
