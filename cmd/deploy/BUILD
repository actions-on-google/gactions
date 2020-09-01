load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/cmd/deploy
gazelle(name = "gazelle")

go_library(
    name = "deploy",
    srcs = ["deploy.go"],
    importpath = "github.com/actions-on-google/gactions/cmd/deploy",
    deps = [
        "//api:sdk",
        "//log",
        "//project",
        "//project:studio",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)
