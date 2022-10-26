.PHONY: protoc test

# make sure we turn on go modules
export GO111MODULE := on

# PROTOC_FLAGS := -I=.. -I=./vendor -I=$(GOPATH)/src
PROTOC_FLAGS := -I=.. -I=$(GOPATH)/src

test:
	go test .

protoc:
#	@go mod vendor
	protoc --gocosmos_out=plugins=interfacetype+grpc,Mgoogle/protobuf/any.proto=github.com/gogo/protobuf/types:. $(PROTOC_FLAGS) ../proofs.proto

install-proto-dep:
	@echo "Installing protoc-gen-gocosmos..."
	@go install github.com/regen-network/cosmos-proto/protoc-gen-gocosmos


