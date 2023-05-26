go_prefix = "github.com/dashjay/mini-lsm-go"

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)

maybe = _maybe

def platform_genrule(*names, suffix = "file"):
    for name in names:
        native.genrule(
            name = name,
            srcs = select({
                "//build:darwin_arm64": ["@%s_darwin_arm64//%s" % (name, suffix)],
                "//build:darwin_amd64": ["@%s_darwin_amd64//%s" % (name, suffix)],
                "//conditions:default": ["@%s_linux_amd64//%s" % (name, suffix)],
            }),
            outs = ["%s_file" % name],
            cmd = "cp $< $@",
            executable = True,
            visibility = ["//visibility:public"],
        )
