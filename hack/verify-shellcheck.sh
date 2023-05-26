#!/bin/bash

# Verify shell by using https://github.com/koalaman/shellcheck

set -o errexit
set -o pipefail

if [[ -n "${TEST_WORKSPACE:-}" ]]; then # Running inside bazel
  echo "Verifying shellcheck in bazel..." >&2
elif ! command -v bazel &> /dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel test --test_output=streamed --test_timeout=1200 //hack:verify-shellcheck
  )
  exit 0
fi

trap 'echo ERROR: shellcheck failed >&2' ERR

shellcheck="$1"

find . -name '*.sh' -not -path '*/node_modules/*' -not -path './bazel-*/*' -exec "$shellcheck" --color=always {} +
