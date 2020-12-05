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
        "//cmd/versions",
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
