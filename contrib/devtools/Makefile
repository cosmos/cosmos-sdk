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
GOLANGCI_LINT_VERSION := v1.16.0
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

TOOLS_DESTDIR  ?= $(GOPATH)/bin

GOLANGCI_LINT   = $(TOOLS_DESTDIR)/golangci-lint
STATIK          = $(TOOLS_DESTDIR)/statik
GOIMPORTS       = $(TOOLS_DESTDIR)/goimports
CLOG            = $(TOOLS_DESTDIR)/clog

all: tools

tools: tools-stamp
tools-stamp: $(STATIK) $(GOIMPORTS) $(CLOG) $(GOLANGCI_LINT)
	touch $@

$(GOLANGCI_LINT): $(mkfile_dir)/install-golangci-lint.sh
	bash $(mkfile_dir)/install-golangci-lint.sh $(TOOLS_DESTDIR) $(GOLANGCI_LINT_VERSION) $(GOLANGCI_LINT_HASHSUM)

$(STATIK):
	$(call go_install,rakyll,statik,v0.1.5)

$(GOIMPORTS):
	go get golang.org/x/tools/cmd/goimports@v0.0.0-20190114222345-bf090417da8b

$(CLOG):
	$(call go_install,alessio,clog,1)

tools-clean:
	rm -f $(STATIK) $(GOIMPORTS) $(CLOG) $(GOLANGCI_LINT)
	rm -f tools-stamp

.PHONY: all tools tools-clean
