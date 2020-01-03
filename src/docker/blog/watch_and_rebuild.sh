#!/usr/bin/env bash
set -euo pipefail

PORT="$1"

POSTS_DIR='posts'

# Cleans up all background tasks
# https://stackoverflow.com/a/2173421/30900
function cleanup_on_exit {
  echo "Cleaning up... "
  # kill 0 kills all members of the process group of the caller.
  # Meaning it will kill this script and all of its background tasks
  # and subshells. https://unix.stackexchange.com/a/67552/179300
  kill 0
}

trap 'exit $?' SIGINT SIGTERM
trap cleanup_on_exit EXIT

function watch() {
  inotifywait --event close_write,moved_to,createDefault \
      --monitor --format '%e %f' "${POSTS_DIR}" |
      while read -r events filename; do
        echo "Detected change. events=${events}, filename=${filename}"
        make html
      done
}

watch &
firebase serve --host 0.0.0.0 --port "${PORT}"

