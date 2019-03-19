###
# Find OS and Go environment
# GO contains the Go binary
# FS contains the OS file separator
###
ifeq ($(OS),Windows_NT)
  GO := $(shell where go.exe 2> NUL)
  FS := \\
else
  GO := $(shell command -v go 2> /dev/null)
  FS := /
endif

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

GOPATH ?= $(shell $(GO) env GOPATH)
GITHUBDIR := $(GOPATH)$(FS)src$(FS)github.com
GOLANGCI_LINT_VERSION := v1.15.0
GOLANGCI_LINT_HASHSUM := ac897cadc180bf0c1a4bf27776c410debad27205b22856b861d41d39d06509cf

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

go_install = $(call go_get,$(1),$(2),$(3)) && cd $(GITHUBDIR)$(FS)$(1)$(FS)$(2) && $(GO) install

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(shell cd $(shell dirname $(mkfile_path)); pwd)

###
# tools
###
all: tools

tools: tools-stamp
tools-stamp: $(GOBIN)/golangci-lint $(GOBIN)/statik $(GOBIN)/goimports $(GOBIN)/gosum $(GOBIN)/sdkch
	touch $@

$(GOBIN)/golangci-lint: contrib/install-golangci-lint.sh $(GOBIN)/gosum
	bash contrib/install-golangci-lint.sh $(GOBIN) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)

$(GOBIN)/statik:
	$(call go_install,rakyll,statik,v0.1.5)

$(GOBIN)/goimports:
	go get golang.org/x/tools/cmd/goimports@v0.0.0-20190114222345-bf090417da8b

$(GOBIN)/gosum:
	go install -mod=readonly ./cmd/gosum/

$(GOBIN)/sdkch:
	go install -mod=readonly ./cmd/sdkch/

tools-clean:
	cd $(GOBIN) && rm -f golangci-lint statik goimports gosum sdkch
	rm -f tools-stamp

.PHONY: all tools tools-clean
