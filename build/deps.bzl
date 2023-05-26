load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load(":consts.bzl", "GOPROXY")

EXPORT_WORKSPACE_IN_BUILD_BAZEL_FILE = [
    "test -f BUILD.bazel && chmod u+w BUILD.bazel || true",
    "echo >> BUILD.bazel",
    "echo 'exports_files([\"WORKSPACE\"], visibility = [\"//visibility:public\"])' >> BUILD.bazel",
]

EXPORT_ALL = """exports_files(glob(["**"]), visibility=["//visibility:public"])"""

def deps():
    http_archive(
        name = "bazel_gazelle",
        sha256 = "ecba0f04f96b4960a5b250c8e8eeec42281035970aa8852dda73098274d14a1d",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.29.0/bazel-gazelle-v0.29.0.tar.gz",
            "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.29.0/bazel-gazelle-v0.29.0.tar.gz",
        ],
    )
    http_archive(
        name = "io_k8s_repo_infra",
        strip_prefix = "repo-infra-0.2.3",
        sha256 = "23d93e6e6ef656661d36b2afd301d277692ded016abe558650b4c813c7c369cf",
        urls = [
            "https://github.com/kubernetes/repo-infra/archive/v0.2.3.tar.gz",
        ],
    )
    http_archive(
        name = "io_bazel_rules_go",
        strip_prefix = "github.com/bazelbuild/rules_go@v0.39.0",
        urls = ["{}/github.com/bazelbuild/rules_go/@v/v0.39.0.zip".format(GOPROXY)],
    )
    http_archive(
        name = "bazel_skylib",
        sha256 = "b8a1527901774180afc798aeb28c4634bdccf19c4d98e7bdd1ce79d1fe9aaad7",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.4.1/bazel-skylib-1.4.1.tar.gz",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.1/bazel-skylib-1.4.1.tar.gz",
        ],
    )
    http_archive(
        name = "com_google_protobuf",
        sha256 = "4ded24230583913b5206a0a20d27b7d19b357bf466bd5cdac994ce3c1c8cbc84",
        strip_prefix = "github.com/protocolbuffers/protobuf@v3.14.0+incompatible",
        type = "zip",
        urls = [
            "{}/github.com/protocolbuffers/protobuf/@v/v3.14.0+incompatible.zip".format(GOPROXY),
        ],
    )
    http_archive(
        name = "rules_python",
        urls = [
            "https://github.com/bazelbuild/rules_python/archive/0.6.0.tar.gz",
        ],
        strip_prefix = "rules_python-0.6.0",
        sha256 = "a30abdfc7126d497a7698c29c46ea9901c6392d6ed315171a6df5ce433aa4502",
    )
    http_archive(
        name = "zlib",
        urls = ["{}/github.com/madler/zlib/@v/v1.2.11.zip".format(GOPROXY)],
        sha256 = "9355229ce4879fe2cfb2ff6ca835076fef41f8c8f9df5f390caef17dcbc2b924",
        build_file_content = """
licenses(["notice"])  #  BSD/MIT-like license

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
    visibility = ["//third_party:__pkg__"],
)

filegroup(
    name = "embedded_tools",
    srcs = glob(["*.c"]) + glob(["*.h"]) + ["BUILD"] + ["LICENSE.txt"],
    visibility = ["//visibility:public"],
)

cc_library(
    name = "zlib",
    srcs = glob(["*.c"]),
    hdrs = glob(["*.h"]),
    # Use -Dverbose=-1 to turn off zlib's trace logging. (#3280)
    copts = [
        "-w",
        "-Dverbose=-1",
    ],
    includes = ["."],
    visibility = ["//visibility:public"],
)""",
    )
    http_archive(
        name = "rules_proto",
        patch_cmds = EXPORT_WORKSPACE_IN_BUILD_BAZEL_FILE,
        sha256 = "8e7d59a5b12b233be5652e3d29f42fba01c7cbab09f6b3a8d0a57ed6d1e9a0da",
        strip_prefix = "rules_proto-7e4afce6fe62dbff0a4a03450143146f9f2d7488",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/rules_proto/archive/7e4afce6fe62dbff0a4a03450143146f9f2d7488.tar.gz",
            "https://github.com/bazelbuild/rules_proto/archive/7e4afce6fe62dbff0a4a03450143146f9f2d7488.tar.gz",
        ],
    )
    http_archive(
        name = "rules_proto_grpc",
        urls = [
            "https://github.com/rules-proto-grpc/rules_proto_grpc/archive/4.1.1.tar.gz",
        ],
        sha256 = "507e38c8d95c7efa4f3b1c0595a8e8f139c885cb41a76cab7e20e4e67ae87731",
        strip_prefix = "rules_proto_grpc-4.1.1",
    )

    # see https://github.com/rules-proto-grpc/rules_proto_grpc/blob/4.1.1/repositories.bzl#L27
    # we mirror it into deploy.i internal network
    http_archive(
        name = "com_github_grpc_grpc",
        urls = [
            "https://github.com/grpc/grpc/archive/v1.42.0.tar.gz",
        ],
        sha256 = "b2f2620c762427bfeeef96a68c1924319f384e877bc0e084487601e4cc6e434c",
        strip_prefix = "grpc-1.42.0",
    )
    http_archive(
        name = "golangci-lint_darwin_amd64",
        sha256 = "e57f2599de73c4da1d36d5255b9baec63f448b3d7fb726ebd3cd64dabbd3ee4a",
        strip_prefix = "golangci-lint-1.52.2-darwin-amd64",
        urls = [
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-darwin-amd64.tar.gz",
        ],
        build_file_content = EXPORT_ALL,
    )

    http_archive(
        name = "golangci-lint_linux_amd64",
        strip_prefix = "golangci-lint-1.52.2-linux-386",
        sha256 = "b2249e43e1624486398f41700dbe4094a4222bf50b2b1b3a740323adb9a1b66f",
        urls = [
            "https://github.com/golangci/golangci-lint/releases/download/v1.52.2/golangci-lint-1.52.2-linux-386.tar.gz",
        ],
        build_file_content = EXPORT_ALL,
    )
    http_archive(
        name = "shellcheck_linux_amd64",
        strip_prefix = "shellcheck-v0.8.0",
        sha256 = "ab6ee1b178f014d1b86d1e24da20d1139656c8b0ed34d2867fbb834dad02bf0a",
        urls = [
            "https://github.com/koalaman/shellcheck/releases/download/v0.8.0/shellcheck-v0.8.0.linux.x86_64.tar.xz",
        ],
        build_file_content = EXPORT_ALL,
    )
    http_archive(
        name = "shellcheck_darwin_amd64",
        strip_prefix = "shellcheck-v0.8.0",
        sha256 = "e065d4afb2620cc8c1d420a9b3e6243c84ff1a693c1ff0e38f279c8f31e86634",
        urls = [
            "https://github.com/koalaman/shellcheck/releases/download/v0.8.0/shellcheck-v0.8.0.darwin.x86_64.tar.xz",
        ],
        build_file_content = EXPORT_ALL,
    )
