FROM bufbuild/buf:latest as BUILDER

FROM golang:alpine

ENV GOLANG_PROTOBUF_VERSION=1.3.5 \
  GOGO_PROTOBUF_VERSION=1.3.2 \
  GRPC_GATEWAY_VERSION=1.14.7 


RUN GO111MODULE=on go get \
  github.com/golang/protobuf/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} \
  github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} \
  github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION} \
  github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${GRPC_GATEWAY_VERSION} \
  github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest

RUN GO111MODULE=on go get -u github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc

COPY --from=BUILDER /usr/local/bin /usr/local/bin
