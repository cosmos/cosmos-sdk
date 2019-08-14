###
# Find OS and Go environment
# GO contains the Go binary
# FS contains the OS file separator
###
ifeq ($(OS),Windows_NT)
  GO := $(shell where go.exe 2> NUL)
  FS := "\\"
else
  GO := $(shell command -v go 2> /dev/null)
  FS := "/"
endif

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

GOPATH ?= $(shell $(GO) env GOPATH)
GITHUBDIR := $(GOPATH)$(FS)src$(FS)github.com
GOLANGCI_LINT_VERSION := v1.17.1
GOLANGCI_LINT_HASHSUM := f5fa647a12f658924d9f7d6b9628d505ab118e8e049e43272de6526053ebe08d

###
# Functions
###

go_get = $(if $(findstring Windows_NT,$(OS)),\
IF NOT EXIST $(GITHUBDIR)$(FS)$(1)$(FS) ( mkdir $(GITHUBDIR)$(FS)$(1) ) else (cd .) &\
IF NOT EXIST $(GITHUBDIR)$(FS)$(1)$(FS)$(2)$(FS) ( cd $(GITHUBDIR)$(FS)$(1) && git clone https://github.com/$(1)/$(2) ) else (cd .) &\
,\
mkdir -p $(GITHUBDIR)$(FS)$(1) &&\
(test ! -d $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && cd $(GITHUBDIR)$(FS)$(1) && git clone https://github.com/$(1)/$(2)) || true &&\
)\
cd $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && git fetch origin && git checkout -q $(3)

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell cd $(shell dirname $(mkfile_path)); pwd)

###
# tools
###

TOOLS_DESTDIR  ?= $(GOPATH)/bin

GOLANGCI_LINT = $(TOOLS_DESTDIR)/golangci-lint
STATIK        = $(TOOLS_DESTDIR)/statik
GOIMPORTS     = $(TOOLS_DESTDIR)/goimports
RUNSIM			  = $(TOOLS_DESTDIR)/runsim

all: tools

tools: statik runsim goimports golangci-lint

golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(mkfile_dir)/install-golangci-lint.sh
	bash $(mkfile_dir)/install-golangci-lint.sh $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)

statik: $(STATIK)
$(STATIK):
	@echo "Get statik"
	@go get github.com/rakyll/statik

goimports: $(GOIMPORTS)
$(GOIMPORTS):
	@echo "Get goimports@v0.0.0-20190628034336-212fb13d595e"
	@go get golang.org/x/tools/cmd/goimports@v0.0.0-20190628034336-212fb13d595e

runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Get runsim"
	@go get github.com/cosmos/tools/cmd/runsim

tools-clean:
	rm -f $(STATIK) $(GOIMPORTS) $(GOLANGCI_LINT) $(RUNSIM)
	rm -f tools-stamp

.PHONY: all tools tools-clean
