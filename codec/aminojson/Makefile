GOLANG_PROTOBUF_VERSION = 1.28.1
BIN = $(CURDIR)/bin
BUF = $(BIN)/buf

.PHONY: proto-gen

$(BIN)/protoc-gen-go:
	GOBIN=$(BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@v$(GOLANG_PROTOBUF_VERSION)

$(BUF):
	./install-buf.sh

proto-gen: $(BUF) $(BIN)/protoc-gen-go
	@echo "Generating proto files"
	PATH=$(BIN) buf generate internal

clean:
	rm -rf $(BIN)