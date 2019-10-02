#!/usr/bin/env bash
set -euo pipefail

inotifywait -e close_write,moved_to,create -m posts |
  while read -r directory events filename; do
    echo "Detected change in ${directory}, events=${events}, filename=${filename}"
    make build
  done
