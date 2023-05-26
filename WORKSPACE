workspace(name = "com_github_dashjay_mini_lsm_go")

load("//build:deps.bzl", "deps")

deps()

load("//build:repos.bzl", "go_repositories", "patch_go_repositories")

# gazelle:repository_macro build/repos.bzl%go_repositories
go_repositories()

patch_go_repositories()

load("//build:loads.bzl", "configure_go", "go_dependencies", "grpc_dependencies")

configure_go()

go_dependencies()

grpc_dependencies()

load("//build:repos.bzl", "go_repositories")

go_repositories()
