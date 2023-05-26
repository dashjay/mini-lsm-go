#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail


fail() {
  echo "ERROR: $1. Fix with:" >&2
  echo "  bazel run @//hack:update-proto" >&2
  exit 1
}


if [[ -n "${TEST_WORKSPACE:-}" ]]; then # Running inside bazel
  echo "Checking protos for changes..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel test --test_output=streamed //hack:verify-proto -- "$@"
  )
  exit 0
fi

proto_path="$1"/"$2"

ret=0
while IFS= read -r -d '' f; do
  project_proto_path=".${f#"$proto_path"}"
  diff -Naupr "$project_proto_path" "$f" || ret=$?
done <   <(find "$proto_path" -type f -print0)

if [[ ${ret} -eq 0 ]]; then
  echo "up to date."
  exit 0
fi
echo "ERROR: proto gen cannot match. Fix with hack/update-proto.sh" >&2
exit 1
