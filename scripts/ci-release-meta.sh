#!/usr/bin/env bash
# Resolve release tag / version / prerelease for GitHub Actions.
# Env: REF_NAME and/or INPUT_TAG. Writes tag, version, image_tag, prerelease
# to GITHUB_OUTPUT.
set -euo pipefail

tag="${INPUT_TAG:-${REF_NAME:-}}"
if [[ -z "${tag}" ]]; then
  echo "REF_NAME or INPUT_TAG is required" >&2
  exit 1
fi
if [[ -z "${GITHUB_OUTPUT:-}" ]]; then
  echo "GITHUB_OUTPUT is required" >&2
  exit 1
fi

ver="${tag#v}"
{
  echo "tag=${tag}"
  echo "version=${ver}"
  echo "image_tag=v${ver}"
  if [[ "${tag}" == *-* ]]; then
    echo "prerelease=true"
  else
    echo "prerelease=false"
  fi
} >> "${GITHUB_OUTPUT}"
