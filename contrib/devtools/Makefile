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

###############################################################################
###                                Functions                                ###
###############################################################################

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


###############################################################################
###                                 Tools                                   ###
###############################################################################

PREFIX ?= /usr/local
BIN ?= $(PREFIX)/bin
UNAME_S ?= $(shell uname -s)
UNAME_M ?= $(shell uname -m)

GOPATH ?= $(shell $(GO) env GOPATH)
GITHUBDIR := $(GOPATH)$(FS)src$(FS)github.com

BUF_VERSION ?= 0.11.0

TOOLS_DESTDIR  ?= $(GOPATH)/bin
STATIK         = $(TOOLS_DESTDIR)/statik
RUNSIM         = $(TOOLS_DESTDIR)/runsim

tools: tools-stamp
tools-stamp: proto-tools statik runsim
	# Create dummy file to satisfy dependency and avoid
	# rebuilding when this Makefile target is hit twice
	# in a row.
	touch $@

# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
statik: $(STATIK)
$(STATIK):
	@echo "Installing statik..."
	@(cd /tmp && go get github.com/rakyll/statik@v0.1.6)

# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will polute the
# go.{mod, sum} files.
#
# ref: https://github.com/golang/go/issues/30515
runsim: $(RUNSIM)
$(RUNSIM):
	@echo "Installing runsim..."
	@(cd /tmp && go get github.com/cosmos/tools/cmd/runsim@v1.0.0)

PROTOC_VERSION ?= 3.13.0
PROTOC_GRPC_GATEWAY_VERSION = 1.14.7
ifeq ($(UNAME_S),Linux)
  PROTOC_ZIP ?= protoc-3.13.0-linux-x86_64.zip
  PROTOC_GRPC_GATEWAY_BIN ?= protoc-gen-grpc-gateway-v1.14.7-linux-x86_64
endif
ifeq ($(UNAME_S),Darwin)
  PROTOC_ZIP ?= protoc-3.13.0-osx-x86_64.zip
  PROTOC_GRPC_GATEWAY_BIN ?= protoc-gen-grpc-gateway-v1.14.7-darwin-x86_64
endif

proto-tools: buf
ifeq (, $(shell which protoc))
	@echo "Installing protoc compiler..."
	@(cd /tmp; \
	curl -OL "https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VERSION}/${PROTOC_ZIP}"; \
	unzip -o ${PROTOC_ZIP} -d $(PREFIX) bin/protoc; \
	unzip -o ${PROTOC_ZIP} -d $(PREFIX) 'include/*'; \
	rm -f ${PROTOC_ZIP})
else
	@echo "protoc already installed; skipping..."
endif

ifeq (, $(shell which protoc-gen-gocosmos))
	@echo "Installing protoc-gen-gocosmos..."
	@go install github.com/regen-network/cosmos-proto/protoc-gen-gocosmos
else
	@echo "protoc-gen-gocosmos already installed; skipping..."
endif

ifeq (, $(shell which protoc-gen-grpc-gateway))
	@echo "Installing protoc-gen-grpc-gateway..."
	@curl -o "${BIN}/protoc-gen-grpc-gateway" -L "https://github.com/grpc-ecosystem/grpc-gateway/releases/download/v${PROTOC_GRPC_GATEWAY_VERSION}/${PROTOC_GRPC_GATEWAY_BIN}"
	@chmod +x "${BIN}/protoc-gen-grpc-gateway"
else
	@echo "protoc-gen-grpc-gateway already installed; skipping..."
endif

ifeq (, $(shell which protoc-gen-swagger))
	@echo "Installing protoc-gen-swagger..."
	@go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
	@npm install -g swagger-combine
else
	@echo "protoc-gen-grpc-gateway already installed; skipping..."
endif
ifeq ($(UNAME_S),Linux)
	@echo "Installing clang-format..."
	sudo apt-get install clang-format
endif
ifeq ($(UNAME_S),Darwin)
	@echo "Installing clang-format..."
	brew install clang-format
endif

buf: buf-stamp

buf-stamp:
	@echo "Installing buf..."
	@curl -sSL \
    "https://github.com/bufbuild/buf/releases/download/v${BUF_VERSION}/buf-${UNAME_S}-${UNAME_M}" \
    -o "${BIN}/buf" && \
	chmod +x "${BIN}/buf"

	touch $@

tools-clean:
	rm -f $(STATIK) $(GOLANGCI_LINT) $(RUNSIM)
	rm -f tools-stamp proto-tools-stamp buf-stamp

.PHONY: tools-clean statik runsim
