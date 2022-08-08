## To test locally:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run --rm -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

FROM bufbuild/buf:1.1.0 as BUILDER

FROM golang:1.19-alpine

RUN apk add --no-cache \
  nodejs \
  npm \
  git \
  make

ENV GOLANG_PROTOBUF_VERSION=1.28.0 \
  GOGO_PROTOBUF_VERSION=1.3.2 \
  GRPC_GATEWAY_VERSION=1.16.0


RUN go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v${GOLANG_PROTOBUF_VERSION}
RUN go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v${GRPC_GATEWAY_VERSION} \
    github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v${GRPC_GATEWAY_VERSION}

# install all gogo protobuf binaries
RUN git clone https://github.com/regen-network/protobuf.git; \
    cd protobuf; \
    go mod download; \
    make install

# we need to use git clone because we use 'replace' directive in go.mod
# protoc-gen-gocosmos was moved to to in cosmos/gogoproto but pending a migration there.
RUN git clone https://github.com/regen-network/cosmos-proto.git; \
    cd cosmos-proto/protoc-gen-gocosmos; \
    go install .

RUN npm install -g swagger-combine

COPY --from=BUILDER /usr/local/bin /usr/local/bin
