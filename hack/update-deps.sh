#!/bin/bash

# This script is a modified version of Kubernetes k8s.io/repo-infra
# https://github.com/kubernetes/repo-infra/blob/v0.0.1-alpha.1/hack/update-deps.sh
# - change goproxy to our own proxy
# - change repos.bzl%go_repositories to build/repos.bzl%go_repositories
# - add --indirect and --import-k8s flag
# - remove update_bazel function and $buildifier, $kazel

# Update vendor and bazel rules to match go.mod
#
# Usage:
#   update-deps.sh [--patch|--minor] [packages]

set -o nounset
set -o errexit
set -o pipefail

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Updating modules..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-deps -- "$@"
  )
  exit 0
fi

unset GOROOT
go=$(realpath "$1")
PATH=$(dirname "$go"):$PATH
export PATH
gazelle=$(realpath "$2")

shift 2

cd "$BUILD_WORKSPACE_DIRECTORY"
trap 'echo "FAILED" >&2' ERR

export GOPROXY=https://goproxy.cn
mode="${1:-}"
shift || true
case "$mode" in
--import-k8s)
    # --import-k8s import k8s.io/kubernetes and edit the go.mod file
    # since directly go get k8s.io/kubernetes may meet unknown revision v0.0.0 error
    # use scripts in https://github.com/kubernetes/kubernetes/issues/79384#issuecomment-521493597
    VERSION=${1#"v"}
    if [ -z "${1#"v"}" ]; then
        echo "Must specify version!"
        exit 1
    fi

    # change to use $GOPROXY for highest speed
    mapfile -t MODS < <(
        curl -sS "$GOPROXY"/k8s.io/kubernetes/@v/v"$VERSION".mod |
        sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
    )
    for MOD in "${MODS[@]}"; do
        V=$(
            $go mod download -json "${MOD}@kubernetes-${VERSION}" |
            sed -n 's|.*"Version": "\(.*\)".*|\1|p'
        )
        $go mod edit "-replace=${MOD}=${MOD}@${V}"
    done
    $go get "k8s.io/kubernetes@v${VERSION}"
    ;;
--indirect)
    if [[ -z "$*" ]]; then
      "$go" get ./...
    else
      "$go" get "$@"
    fi
    ;;
--minor)
    if [[ -z "$*" ]]; then
      "$go" get -u ./...
    else
      "$go" get -u "$@"
    fi
    ;;
--patch)
    if [[ -z "$*" ]]; then
      "$go" get -u=patch ./...
    else
      "$go" get -u=patch "$@"
    fi
    ;;
"")
    # Just validate, or maybe manual go.mod edit
    ;;
*)
    echo "Usage: $(basename "$0") [--patch|--minor] [packages]" >&2
    exit 1
    ;;
esac

rm -rf vendor
"$go" mod tidy
"$gazelle" update-repos \
  --from_file=go.mod --to_macro=build/repos.bzl%go_repositories \
  --build_file_generation=on --build_file_proto_mode=disable \
  --prune
echo "SUCCESS: updated modules"
