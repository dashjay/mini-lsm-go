#!/usr/bin/env bash

# Update *.pb.go in proto directory defined in BUILD.bazel

set -o errexit
set -o nounset
set -o pipefail

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Generating proto files..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-proto -- "$@"
  )
  exit 0
fi

proto_source_dir="${BUILD_WORKSPACE_DIRECTORY}"/"$1"/"$2"

(
  cd "$proto_source_dir"
  find . -name "*.go" -exec rm -f "${BUILD_WORKSPACE_DIRECTORY}"/{} \;
)

shift 2 || true
mode="${1:-}"
case "$mode" in
--remove)
    exit 0
esac

cp -r "$proto_source_dir"/* "${BUILD_WORKSPACE_DIRECTORY}"/
