# handcrafted

[![ci](https://github.com/WillAbides/handcrafted/workflows/ci/badge.svg?branch=master&event=push)](https://github.com/WillAbides/handcrafted/actions?query=workflow%3Aci+branch%3Amaster+event%3Apush)

handcrafted is a simple command-line tool to help you list handcrafted go files (files that aren't generated). It
 takes a list of files in stdin, and outputs the matching files from the list to stdout.

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

```shell script
filter="{{ range .GoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .TestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}{{ range .XTestGoFiles }}{{$.Dir}}/{{.}}
{{ end }}"

go list -f "$filter" runtime/... | handcrafted -generated
```

Create `script/fmt` that runs goimports only on handcrafted files:

```shell script
#!/bin/sh

set -e

cd "$(git rev-parse --show-toplevel)"

if ! [ -f "bin/handcrafted" ]; then
  GOBIN="$(pwd)/bin" go get github.com/willabides/handcrafted
fi

git ls-files -o -c --exclude-standard -- *.go |
 bin/handcrafted | 
 xargs goimports -w

```
