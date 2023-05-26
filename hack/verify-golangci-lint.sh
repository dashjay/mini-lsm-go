#!/bin/bash

# This script is a modified version of Kubernetes k8s.io/repo-infra
# https://github.com/kubernetes/repo-infra/blob/v0.0.1-alpha.1/hack/verify-golangci-lint.sh
# - set GOPROXY=https://goproxy.cn and set GOSUMDB=off
# - set --color=always while running golangci-lint
# - only use linux_amd64 for lint
# - add LOCAL_MODE support in local machine for enable cache acceleration

set -o errexit
set -o pipefail

if [[ -n "${TEST_WORKSPACE:-}" ]]; then # Running inside bazel
  echo "Verifying golangci-lint in bazel..." >&2
elif [[ -n "${LOCAL_MODE}" ]]; then
  echo "Verifying golangci-lint in local mode..." >&2
  currentDir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
  cd "$currentDir"/.. # cd rootDir
elif ! command -v bazel &> /dev/null; then
  echo "Install bazel by using ./hack/install-bazel.sh" >&2
  exit 1
else
  (
    set -o xtrace
    bazel test --test_output=streamed --test_timeout=1200 //hack:verify-golangci-lint
  )
  exit 0
fi

trap 'echo ERROR: golangci-lint failed >&2' ERR

if [[ ! -f .golangci.yml ]]; then
  echo 'ERROR: missing .golangci.yml in repo root' >&2
  exit 1
fi

if [[ -n "$LOCAL_MODE" ]]; then
  echo -e "[INFO] \033[33mGetting bazel info...\033[0m"
  bazel_bin=$(bazel info bazel-bin)
  output_base=$(bazel info output_base)
  bazel build //build:golangci-lint @go_sdk//:bin/go
  go="$output_base"/external/go_sdk/bin/go
  golangci_lint="$bazel_bin"/build/golangci-lint_file
  go_bin=$(dirname "$go")
  GOROOT=$(dirname "$go_bin")
else
  go="$1"
  golangci_lint=$2
  shift 2
  export HOME=$TEST_TMPDIR/home
fi

export GO111MODULE=on
# change proxy to inner https://goproxy.cn
export GOPROXY=${GOPROXY:-https://goproxy.cn}
export GOSUMDB=${GOSUMDB:-off}

export GOPATH=$HOME/go

go=$(realpath "$go")
PATH=$(dirname "$go"):$PATH
export PATH

echo -e "[INFO] \033[33mGOPATH=$GOPATH GOROOT=$GOROOT\033[0m"
echo -e "[INFO] \033[33mGOOS=$GOOS GOARCH=$GOARCH\033[0m"
echo -e "[INFO] \033[33mgo binary is in $go\033[0m"
echo -e "[INFO] \033[33m$("$go" version)\033[0m"
echo -e "[INFO] \033[33mgolangci-lint cache status is:\033[0m"
echo -e "\033[33m$("$golangci_lint" cache status)\033[0m"

"$golangci_lint" run --color=always "$@"
