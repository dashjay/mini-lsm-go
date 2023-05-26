#!/usr/bin/env bash

# This script is a modified version of Kubernetes k8s.io/repo-infra
# https://github.com/kubernetes/repo-infra/blob/v0.0.1-alpha.1/hack/update-bazel.sh
# we remove go.mod check and add exclusion of '*/third_party/*' while using $buildifier

set -o errexit
set -o nounset
set -o pipefail


fail() {
  echo "ERROR: $1. Fix with:" >&2
  echo "  bazel run //hack:update-bazel" >&2
  exit 1
}

if [[ -n "${TEST_WORKSPACE:-}" ]]; then # Running inside bazel
  echo "Validating bazel rules..." >&2
elif ! command -v bazel &> /dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
elif ! bazel query @//:all-srcs &>/dev/null; then
  fail "bazel rules need bootstrapping"
else
  (
    set -o xtrace
    bazel test --test_output=streamed //hack:verify-bazel
  )
  exit 0
fi

buildifier=$1
gazelle=$2
kazel=$3

gazelle_diff=$("$gazelle" fix --mode=diff --external=external || echo "ERROR: gazelle diffs")
kazel_diff=$("$kazel" --dry-run --print-diff --cfg-path=./.kazelcfg.json || echo "ERROR: kazel diffs")
# TODO(fejta): --mode=diff --lint=warn
buildifier_diff=$(find . \
  -name BUILD -o -name BUILD.bazel -o -name '*.bzl' -type f \
  -exec "$buildifier" --mode=diff '{}' + 2>&1 || echo "ERROR: found buildifier diffs")

if [[ -n "${gazelle_diff}${kazel_diff}${buildifier_diff}" ]]; then
  echo "Current rules (-) do not match expected (+):" >&2
  echo "gazelle diff:"
  echo "${gazelle_diff}"
  echo "kazel diff:"
  echo "${kazel_diff}"
  echo "buildifier diff:"
  echo "$buildifier_diff"
  echo
  fail "bazel rules out of date"
fi

