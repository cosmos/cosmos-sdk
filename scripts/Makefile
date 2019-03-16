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
tools: $(GOBIN)/golangci-lint $(GOBIN)/statik $(GOBIN)/goimports $(GOBIN)/gosum $(GOBIN)/sdkch

#v2.0.11
$(GOBIN)/golangci-lint:
	sh tools/install-golangci.sh -d -b $(GOBIN)

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

.PHONY: all tools-clean
