# Copyright (2021) Cobalt Speech and Language Inc.

# Needed tools are installed to BINDIR.
BINDIR := ./tmp/bin

LINTER := $(BINDIR)/golangci-lint
LINTER_VERSION := 1.30.0

# Linux vs Darwin detection for the machine on which the build is taking place (not to be used for the build target)
DEV_OS := $(shell uname -s | tr A-Z a-z)

$(LINTER):
	mkdir -p $(BINDIR)
	wget "https://github.com/golangci/golangci-lint/releases/download/v$(LINTER_VERSION)/golangci-lint-$(LINTER_VERSION)-$(DEV_OS)-amd64.tar.gz" -O - | tar -xz -C $(BINDIR) --strip-components=1 --exclude=README.md --exclude=LICENSE

# Run go-fmt on all go files.  We list all go files in the repository, run
# gofmt.  gofmt produces output with a list of files that have fmt errors.  If
# we have an empty output, we exit with 0 status, otherwise we exit with nonzero
# status.
.PHONY: fmt-check
fmt-check:
	BADFILES=$$(gofmt -l -d $$(find . -type f -name '*.go')) && [ -z "$$BADFILES" ] && exit 0

# Run go-fmt and automatically fix issues
.PHONY: fmt
fmt:
	gofmt -s -w $$(find . -type f -name '*.go')

# Run lint checks
.PHONY: lint-check
lint-check: $(LINTER)
	$(LINTER) run

# Run tests
.PHONY: test
test:
	go test -race -cover ./...

# Nothing to build
.PHONY: build
build:
	echo "Nothing to build"
