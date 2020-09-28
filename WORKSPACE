load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "2697f6bc7c529ee5e6a2d9799870b9ec9eaeb3ee7d70ed50b87a2c2c97e13d9e",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.23.8/rules_go-v0.23.8.tar.gz",
    ],
)

load("@io_bazel_rules_go//extras:embed_data_deps.bzl", "go_embed_data_dependencies")
load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

go_embed_data_dependencies()

# Import Bazel Gazelle
http_archive(
    name = "bazel_gazelle",
    sha256 = "d4113967ab451dd4d2d767c3ca5f927fec4b30f3b2c6f8135a2033b9c05a5687",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.0/bazel-gazelle-v0.22.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.22.0/bazel-gazelle-v0.22.0.tar.gz",
    ],
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")
gazelle_dependencies()

go_repository(
    name = "com_github_fatih_color",
    commit = "daf2830f2741ebb735b21709a520c5f37d642d85",
    importpath = "github.com/fatih/color",
)

go_repository(
    name = "com_github_golang_crypto",
    commit = "5c72a883971a4325f8c62bf07b6d38c20ea47a6a",
    importpath = "github.com/golang/crypto",
)

go_repository(
    name = "com_github_google_go_cmp",
    commit = "d2fcc899bdc2d134b7c00e36137260db963e193c",
    importpath = "github.com/google/go-cmp",
)

go_repository(
    name = "com_github_google_uuid",
    commit = "0e4e31197428a347842d152773b4cace4645ca25",
    importpath = "github.com/google/uuid",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    commit = "76626ae9c91c4f2a10f34cad8ce83ea42c93bb75",
    importpath = "github.com/inconshreveable/mousetrap"
)

go_repository(
    name = "com_github_pborman_uuid",
    commit = "5b6091a6a160ee5ce12917b21ab96acec2a4fdc0",
    importpath = "github.com/pborman/uuid",
)

go_repository(
    name = "com_github_protolambda_messagediff",
    commit = "24215fdae608ad3b414abaa5b08c54386bf6c774",
    importpath = "github.com/protolambda/messagediff",
)

go_repository(
    name = "com_github_spf13_cobra",
    commit = "02a0d2fbc9e61d26f8e5979749f6030964a55a3e",
    importpath = "github.com/spf13/cobra",
)

go_repository(
    name = "com_github_spf13_pflag",
    commit = "81378bbcd8a1005f72b1e8d7579e5dd7b2d612ab",
    importpath = "github.com/spf13/pflag",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "aa4dea45b99b7440f266638cbd3d8d9504e93bd7",
    importpath = "cloud.google.com/go",
    remote = "https://github.com/googleapis/google-cloud-go",
    vcs = "git",
)

go_repository(
    name = "in_gopkg_yaml",
    tag = "v2.3.0",
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "org_golang_google_appengine",
    commit = "5539592",
    importpath = "google.golang.org/appengine",
    remote = "https://github.com/golang/appengine",
    vcs = "git",
)

go_repository(
    name = "org_golang_x_net",
    commit = "1e06a53dbb7e2ed46e91183f219db23c6943c532", # v0.0.0-20190108225652-1e06a53dbb7e as specified in oauth2,
    importpath = "golang.org/x/net",
    remote = "https://github.com/golang/net",
    vcs = "git",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "bf48bf16ab8d622ce64ec6ce98d2c98f916b6303",
    importpath = "golang.org/x/oauth2",
    remote = "https://github.com/golang/oauth2",
    vcs = "git",
)
