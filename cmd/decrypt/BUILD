load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/cmd/decrypt
gazelle(name = "gazelle")

test_suite(
    name = "all_tests",
    tags = ["-notwindows"],
)

go_library(
    name = "decrypt",
    srcs = ["decrypt.go"],
    importpath = "github.com/actions-on-google/gactions/cmd/decrypt",
    deps = [
        "//api:sdk",
        "//log",
        "//project",
        "@com_github_spf13_cobra//:go_default_library",
        "@in_gopkg_yaml//:go_default_library",
    ],
)

go_test(
    name = "decrypt_test",
    size = "small",
    srcs = ["decrypt_test.go"],
    embed = [":decrypt"],
)
