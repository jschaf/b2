#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(git rev-parse --show-toplevel)"
# Match the version used by KaTeX in:
# https://github.com/graemephi/goldmark-qjs-katex/tree/master/katex/katex
katex_version='0.16.11'

# Update the compiled CSS.
echo 'update css'
css_url="https://cdn.jsdelivr.net/npm/katex@$katex_version/dist/katex.min.css"
curl --fail --location --output "$repo_dir/style/katex.min.css" "$css_url"

# Update fonts.
echo 'update fonts'
archive_url="https://github.com/KaTeX/KaTeX/archive/refs/tags/v$katex_version.zip"
dir="$(mktemp -d)"
curl --fail --location --output "$dir/katex.zip" "$archive_url"

# Extract font files from the archive
rm -f "$repo_dir/style/fonts/KaTeX_"*
unzip -j "$dir/katex.zip" "KaTeX-$katex_version/fonts/*" -d "$repo_dir/style/fonts/"

# Cleanup temporary directory
rm -rf "$dir"
