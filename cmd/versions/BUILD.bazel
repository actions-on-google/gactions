load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/cmd/pull
gazelle(name = "gazelle")

go_library(
    name = "versions",
    srcs = ["versions.go"],
    importpath = "github.com/actions-on-google/gactions/cmd/versions",
    deps = [
        "//api:sdk",
        "//project",
        "//project:studio",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)
