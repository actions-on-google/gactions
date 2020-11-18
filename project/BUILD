load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")
load("@bazel_gazelle//:def.bzl", "gazelle")

package(default_visibility = ["//visibility:public"])

# gazelle:prefix github.com/actions-on-google/gactions/project
gazelle(name = "gazelle")

test_suite(
    name = "all_tests",
    tags = ["-notwindows"],
)

go_library(
    name = "project",
    srcs = ["project.go"],
    importpath = "github.com/actions-on-google/gactions/project",
)

go_library(
    name = "studio",
    srcs = ["studio.go"],
    importpath = "github.com/actions-on-google/gactions/project/studio",
    deps = [
        ":project",
        "//api:yamlutils",
        "@in_gopkg_yaml//:go_default_library",
        "//log",
    ],
)

go_test(
    name = "studio_test",
    size = "small",
    srcs = ["studio_test.go"],
    embed = [":studio"],
    deps = [
        ":project",
        "//api:testutils",
        "@com_github_google_go_cmp//cmp:go_default_library",
    ],
)
