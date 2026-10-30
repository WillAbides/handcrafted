# handcrafted

[![ci](https://github.com/WillAbides/handcrafted/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/WillAbides/handcrafted/actions/workflows/ci.yaml)

handcrafted is a simple command-line tool to help you list handcrafted go files (files that aren't generated). It
takes a list of files in stdin, and outputs the matching files from the list to stdout.

Files are classified using Go's standard generated-code marker, a comment of
the form `// Code generated ... DO NOT EDIT.` before the `package` clause.
Files without a valid package clause are omitted from both outputs. Unreadable
files cause an error.

## Installation

Download a binary from [GitHub Releases](https://github.com/WillAbides/handcrafted/releases),
or install with Go:

```sh
go install github.com/willabides/handcrafted@latest
```

## Usage

Get all non-generated files in the current directory:

`ls *.go | handcrafted`

Add `-generated` flag to show generated files instead:

`ls *.go | handcrafted -generated`

Use with `find` to do the same recursively:

`find . -name "*.go" | handcrafted`

Use with `git ls-files` to list all handcrafted files in the git repo

`git ls-files -- *.go | handcrafted`

Use with `go list` to get generated files from an arbitrary package:

```sh
filter="{{ range .GoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .TestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .XTestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}"

go list -f "$filter" runtime/... | handcrafted -generated
```

Create `script/fmt` that runs goimports only on handcrafted files:

```sh
#!/bin/sh

set -e

cd "$(git rev-parse --show-toplevel)"

if ! [ -f "bin/handcrafted" ]; then
  GOBIN="$(pwd)/bin" go install github.com/willabides/handcrafted@latest
fi

git ls-files -o -c --exclude-standard '*.go' |
  bin/handcrafted |
  xargs goimports -w
```

## Development

Use the Go version in `go.mod`. Development tools are pinned in `.bindown.yaml`
and installed automatically by `script/bindown`.

| Command | Purpose |
| --- | --- |
| `script/handcrafted` | Build and run the local CLI |
| `script/fmt` | Format handcrafted Go files and shell scripts |
| `script/test` | Run tests with race detection and coverage |
| `script/lint` | Run golangci-lint and ShellCheck |
| `script/generate` | Run Go generators and `go mod tidy` |
| `script/generate --check` | Check generation in an isolated worktree, including uncommitted changes |
| `script/cibuild` | Run CI checks and build snapshot binaries |

Pass Go test flags to `script/test`, for example `script/test -run TestScript`.
CLI scenarios are [testscript](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript)
archives in `testdata/script`.

CI checks pull requests and pushes to `main`. Release-train validates release
labels on pull requests and publishes releases from `main` only after CI passes.
`script/release-hook` packages binaries with GoReleaser for release-train to
upload. Releases also include a Bindown dependency template.
