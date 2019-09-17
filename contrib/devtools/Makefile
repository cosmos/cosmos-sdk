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
GOLANGCI_LINT_VERSION := v1.18.0
GOLANGCI_LINT_HASHSUM := 8d21cc95da8d3daf8321ac40091456fc26123c964d7c2281d339d431f2f4c840

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
SWAG          = $(TOOLS_DESTDIR)/swag
RUNSIM			  = $(TOOLS_DESTDIR)/runsim

all: tools

tools: swag runsim golangci-lint

golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(mkfile_dir)/install-golangci-lint.sh
	bash $(mkfile_dir)/install-golangci-lint.sh $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)

swag: $(SWAG)
$(SWAG):
	@echo "Get swag"
	@go get -u github.com/swaggo/swag/cmd/swag

runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Get runsim"
	@go get github.com/cosmos/tools/cmd/runsim

tools-clean:
	rm -f $(SWAG) $(GOLANGCI_LINT) $(RUNSIM)
	rm -f tools-stamp

.PHONY: all tools tools-clean
