load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_embed_data", "go_library")
load("//versions:version_names.bzl", "APP_VERSION")
load("@bazel_gazelle//:def.bzl", "gazelle")

# gazelle:prefix github.com/actions-on-google/gactions
gazelle(name = "gazelle")

test_suite(
    name = "all_tests",
    tags = ["-notwindows"],
    visibility = [":__pkg__"],
)

go_binary(
    name = "gactions",
    srcs = ["gactions.go"],
    # https://github.com/bazelbuild/rules_go/blob/master/go/core.rst#defines-and-stamping
    x_defs = {
        'gactions.version': APP_VERSION,
    },
    deps = [":cli"],
)

genrule(
    name = "package_tar",
    srcs = [":gactions"],
    outs = [
        "aog_cli.tar.gz",
        "aog_cli.tar.gz.sha256",
    ],
    cmd = "&&".join([
        "mkdir aog_cli",
        "mv $(SRCS) aog_cli/",
        "tar czvf $(@D)/aog_cli.tar.gz --owner=0 --group=0 --numeric-owner --mtime=@0 --dereference aog_cli",
        "cd $(@D) && sha256sum aog_cli.tar.gz > aog_cli.tar.gz.sha256",
    ]),
)

genrule(
    name = "package_zip",
    srcs = [":gactions"],
    outs = [
        "aog_cli.zip",
        "aog_cli.zip.sha256",
    ],
    # NOTE: This assumes the output is packaged for windows! If you really want
    # a zip in linux, you're doing it wrong. Use package_tar to get a tarball.
    cmd = "&&".join([
        "mkdir aog_cli",
        "cp -Lr $(SRCS) aog_cli/",
        "find aog_cli -type f \\! -name '*.*' -exec mv {} {}.exe \\;",
        "find aog_cli -exec touch -d '1970-01-01' {} \\;",
        "TZ=UTC zip -X -r $(@D)/aog_cli.zip aog_cli",
        "cd $(@D) && sha256sum aog_cli.zip > aog_cli.zip.sha256",
    ]),
)

go_library(
    name = "cli",
    srcs = [
        "cli.go",
        ":client_not_so_secret_embed_data_go",
    ],
    importpath = "github.com/actions-on-google/gactions/cli",
    deps = [
        "//api:sdk",
        "//cmd/decrypt",
        "//cmd/deploy",
        "//cmd/encrypt",
        "//cmd/ginit",
        "//cmd/login",
        "//cmd/logout",
        "//cmd/notices",
        "//cmd/pull",
        "//cmd/push",
        "//cmd/releasechannels",
        "//cmd/version",
        "//log",
        "//project:studio",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)

go_embed_data(
    name = "client_not_so_secret_embed_data_go",
    src = "data/client_not_so_secret.json",
    package = "cli",
    var = "clientNotSoSecretJSON",
)
