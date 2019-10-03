#!/usr/bin/env bash
set -euo pipefail

# Builds the Docker image for this project.
#
# If none of the dependencies that build the Docker image changed, this script
# halts the CircleCI job that runs this script.

printf '
# Checking if Docker image is stale
===================================\n'

# The hash file persisted in the workspace by the checkout_code job.
current_hash_file='.circleci/docker_cache_key'
if [[ ! -e "${current_hash_file}" ]]; then
  echo "ERROR: cache key file not found at ${current_hash_file}."
  exit 1
fi
current_hash="$(< "${current_hash_file}")"

# The hash file of the already built docker file.
precheck_hash_file=".circleci/docker_precheck_cache_key"
precheck_hash='<none>'
precheck_status='stale'
if [[ -f "${precheck_hash_file}" ]]; then
  precheck_hash="$(< ${precheck_hash_file})"
  if [[ "${current_hash}" == "${precheck_hash}" ]]; then
    precheck_status='fresh'
  fi
fi

printf 'Hashes in %-65s\n%s\n\n' "${current_hash_file}" "${current_hash}"
printf 'Hashes in %-65s\n%s\n\n' "${precheck_hash_file}" "${precheck_hash}"
echo "cache_status: ${precheck_status}"

if [[ "${precheck_status}" == 'fresh' ]]; then
  printf '\nSkipping building Docker image because because the cache is fresh.\n'
  circleci task halt
  exit 0
fi
