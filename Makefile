GOCMD=go
GOBUILD=$(GOCMD) build
PATH := "${CURDIR}/bin:$(PATH)"

.PHONY: gobuildcache

bin/golangci-lint:
	script/bindown install $(notdir $@)

bin/shellcheck:
	script/bindown install $(notdir $@)

bin/handcrafted: gobuildcache
	go build -o $@ .

bin/gobin: bin/bindown
	bin/bindown install $(notdir $@)

GOFUMPT_REV := abc0db2c416aca0f60ea33c23c76665f6e7ba0b6
bin/gofumpt: bin/gobin
	GOBIN=${CURDIR}/bin \
	bin/gobin mvdan.cc/gofumpt@$(GOFUMPT_REV)
