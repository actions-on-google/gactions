load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/cmd/push
gazelle(name = "gazelle")

test_suite(
    name = "all_tests",
    tags = ["-notwindows"],
)

go_library(
    name = "push",
    srcs = ["push.go"],
    importpath = "github.com/actions-on-google/gactions/cmd/push",
    deps = [
        "//api:sdk",
        "//log",
        "//project",
        "//project:studio",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)

go_test(
    name = "push_test",
    size = "small",
    srcs = ["push_test.go"],
    embed = [":push"],
    tags = ["notwindows"],
    deps = [
        "//project",
        "//project:studio",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)
