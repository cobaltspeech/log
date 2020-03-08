# Copyright (2020) Cobalt Speech and Language Inc.

# Needed tools are installed to BINDIR.
BINDIR := ./tmp/bin

LINTER := $(BINDIR)/golangci-lint
LINTER_VERSION := 1.23.8

# Linux vs Darwin detection for the machine on which the build is taking place (not to be used for the build target)
DEV_OS := $(shell uname -s | tr A-Z a-z)

$(LINTER):
	mkdir -p $(BINDIR)
	wget "https://github.com/golangci/golangci-lint/releases/download/v$(LINTER_VERSION)/golangci-lint-$(LINTER_VERSION)-$(DEV_OS)-amd64.tar.gz" -O - | tar -xz -C $(BINDIR) --strip-components=1 --exclude=README.md --exclude=LICENSE

# Run go-fmt on all go files.  We list all go files in the repository, run
# gofmt.  gofmt produces output with a list of files that have fmt errors.  If
# we have an empty output, we exit with 0 status, otherwise we exit with nonzero
# status.
.PHONY: fmt
fmt:
	BADFILES=$$(gofmt -l -d $$(find . -type f -name '*.go')) && [ -z "$$BADFILES" ] && exit 0

# Run lint checks
.PHONY: lint
lint: $(LINTER)
	$(LINTER) run

# Run tests
.PHONY: test
test:
	go test -cover ./...
	go test -cover ./... -tags cobalt_log_trace
