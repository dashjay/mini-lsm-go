load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
load("@io_bazel_rules_go//go:deps.bzl", "go_download_sdk", "go_register_toolchains", "go_rules_dependencies")
load("@rules_proto_grpc//:repositories.bzl", "rules_proto_grpc_repos", "rules_proto_grpc_toolchains")
load("@rules_proto//proto:repositories.bzl", "rules_proto_dependencies", "rules_proto_toolchains")
load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")
load(":utils.bzl", _maybe = "maybe")

def configure_go(version = "1.19.9"):
    go_download_sdk(
        name = "go_sdk",
        version = version,
    )

    go_rules_dependencies()

    go_register_toolchains()

def go_dependencies():
    gazelle_dependencies()

def grpc_dependencies():
    # To build go_proto_library rules with the gRPC plugin, org_golang_google_grpc, org_golang_x_net, and org_golang_x_text is needed
    # see https://github.com/bazelbuild/rules_go/blob/v0.20.1/go/workspace.rst#grpc-dependencies
    _maybe(
        go_repository,
        name = "org_golang_google_grpc",
        build_file_proto_mode = "disable",
        importpath = "google.golang.org/grpc",
        sum = "h1:J0UbZOIrCAl+fpTOf8YLs4dJo8L/owV4LYVtAXQoPkw=",
        version = "v1.22.0",
    )

    _maybe(
        go_repository,
        name = "org_golang_x_net",
        importpath = "golang.org/x/net",
        sum = "h1:oWX7TPOiFAMXLq8o0ikBYfCJVlRHBcsciT5bXOrH628=",
        version = "v0.0.0-20190311183353-d8887717615a",
    )

    _maybe(
        go_repository,
        name = "org_golang_x_text",
        importpath = "golang.org/x/text",
        sum = "h1:g61tztE5qeGQ89tm6NTjjM9VPIm088od1l6aSorWRWg=",
        version = "v0.3.0",
    )

    protobuf_deps()
    rules_proto_grpc_repos()
    rules_proto_grpc_toolchains()
    rules_proto_dependencies()
    rules_proto_toolchains()
