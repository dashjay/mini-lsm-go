#!/usr/bin/env bash

# Cause the script to exit if a single command fails
set -e

platform="unknown"
unamestr="$(uname)"
archstr="$(uname -m)"
arch="x86_64"
if [[ "$unamestr" == "Linux" ]]; then
  platform="linux"
elif [[ "$unamestr" == "Darwin" ]]; then
  platform="darwin"
else
  echo "Unrecognized platform."
  exit 1
fi

if [[ "$archstr" == "arm64" ]] || [[ "$archstr" == "aarch64" ]]; then
  arch="arm64"
fi

VERSION=5.4.0

echo "Platform is ${platform}-${arch}, will install bazel $VERSION"

if test $# -gt 0; then
    VERSION=$1
fi

URL="https://github.com/bazelbuild/bazel/releases/download/${VERSION}/bazel-${VERSION}-installer-${platform}-${arch}.sh"
curl -L -o /tmp/install.sh "$URL"
chmod +x /tmp/install.sh
/tmp/install.sh --user
rm -f /tmp/install.sh
