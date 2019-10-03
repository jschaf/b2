#!/bin/bash
set -euo pipefail

# Creates a combined cache key to use for the CircleCI docker cache.
#
# A clean build of the Docker image depends on more than a single file. CircleCI
# can't use multiple files as a cache key, so store the hashes of all
# dependencies in a single file that we can then reference from a save_cache
# step. See https://discuss.circleci.com/t/base-cache-key-on-checksum-of-directory-rather-than-a-single-file/20059/4.

# The hashes of all dependencies for building the Docker image.
create_cache_key_script_hash="$(sha1sum .circleci/create_docker_cache_key.sh)"
makefile_hash="$(sha1sum Makefile)"
dockerfile_hashes="$(find docker/blog -type f -print0 | xargs -0 sha1sum | sort)"

# Where to save the hashes.
cache_key_file='.circleci/docker_cache_key'

# If you change the logic for either building the Docker image, update
# this key to the current time. It's easier to change it here than changing
# the cache name in .circleci/config.yml.
cache_bust_key='2019-10-03T04:33'

printf '\nWriting current Docker cache key to %s.\n\n' ${cache_key_file}
echo "# Combined cache key hashes
cache_buster=${cache_bust_key}
${create_cache_key_script_hash}
${makefile_hash}
${dockerfile_hashes}" > ${cache_key_file}

cat "${cache_key_file}"
