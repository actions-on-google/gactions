load("@io_bazel_rules_go//go:def.bzl", "go_library")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/cmd/encrypt
gazelle(name = "gazelle")

go_library(
    name = "encrypt",
    srcs = ["encrypt.go"],
    importpath = "github.com/actions-on-google/gactions/cmd/encrypt",
    deps = [
        "//api:sdk",
        "//log",
        "//project",
        "@com_github_golang_crypto//ssh/terminal:go_default_library",
        "@com_github_spf13_cobra//:go_default_library",
    ],
)
