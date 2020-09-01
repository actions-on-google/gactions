load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/log
gazelle(name = "gazelle")

test_suite(
    name = "all_tests",
    tags = ["-notwindows"],
)

go_library(
    name = "log",
    srcs = ["log.go"],
    importpath = "github.com/actions-on-google/gactions/log",
    deps = [
        "@com_github_fatih_color//:go_default_library",
    ],
)
