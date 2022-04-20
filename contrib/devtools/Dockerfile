## To test locally:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

FROM bufbuild/buf:1.1.0 as BUILDER

FROM golang:alpine

RUN apk add --no-cache \
  nodejs \
  npm \
  git \
  make

ENV GOLANG_PROTOBUF_VERSION=1.28.0 \
  GOGO_PROTOBUF_VERSION=1.3.2 \
  GRPC_GATEWAY_VERSION=1.16.0


RUN go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest && \
  go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION} && \
  go install github.com/gogo/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} && \
  go install github.com/gogo/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} && \ 
  go install github.com/gogo/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION} && \
  go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION} && \
  go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${GRPC_GATEWAY_VERSION} && \
  go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@v1.4.1


RUN npm install -g swagger-combine

COPY --from=BUILDER /usr/local/bin /usr/local/bin
