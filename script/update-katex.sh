#!/usr/bin/env bash

set -euo pipefail

repo_dir="$(git rev-parse --show-toplevel)"

# Update the compiled CSS.
css_url='https://cdn.jsdelivr.net/npm/katex@latest/dist/katex.min.css'
curl --fail --location --output "$repo_dir/style/katex.min.css" "$css_url"

# Update fonts.
archive_url='https://github.com/KaTeX/KaTeX/archive/refs/heads/main.zip'
dir="$(mktemp -d)"
curl --fail --location --output "$dir/katex.zip" "$archive_url"

# Extract font files from the archive
rm -f "$repo_dir/style/fonts/KaTeX_"*
unzip -j "$dir/katex.zip" "KaTeX-main/fonts/*" -d "$repo_dir/style/fonts/"

# Cleanup temporary directory
rm -rf "$dir"
