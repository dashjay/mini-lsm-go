#!/usr/bin/env bash

# This script is a modified version of Kubernetes k8s.io/repo-infra
# https://github.com/kubernetes/repo-infra/blob/v0.0.1-alpha.1/hack/update-bazel.sh
# we remove go.mod check and add exclusion of '*/third_party/*' while using $buildifier

set -o errexit
set -o nounset
set -o pipefail

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Updating bazel rules..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-bazel
  )
  exit 0
fi

buildifier=$(realpath "$1")
gazelle=$(realpath "$2")
kazel=$(realpath "$3")

cd "$BUILD_WORKSPACE_DIRECTORY"

set -o xtrace
"$gazelle" fix --external=external
"$kazel" --cfg-path=./.kazelcfg.json
find . -name BUILD -o -name BUILD.bazel -o -name '*.bzl' -type f \
  -exec "$buildifier" --mode=fix --lint=fix '{}' +
