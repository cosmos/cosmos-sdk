## To test locally:
# docker build --pull --rm -f "contrib/devtools/Dockerfile" -t cosmossdk-proto:latest "contrib/devtools"
# docker run -v $(pwd):/workspace --workdir /workspace cosmossdk-proto sh ./scripts/protocgen.sh

FROM bufbuild/buf:1.1.0 as BUILDER

FROM golang:1.18-alpine

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
RUN go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@latest

## For the moment there is a problem with go install and replace, so installing
## by cloninng the repository instead
#RUN go get github.com/regen-network/protobuf/protoc-gen-gogo@v${GOGO_PROTOBUF_VERSION} \
#   github.com/regen-network/protobuf/protoc-gen-gogofast@v${GOGO_PROTOBUF_VERSION} \
#   github.com/regen-network/protobuf/protoc-gen-gogofaster@v${GOGO_PROTOBUF_VERSION}

RUN git clone https://github.com/regen-network/protobuf.git; \
    cd protobuf; \
    go mod download; \
    make install


RUN npm install -g swagger-combine

COPY --from=BUILDER /usr/local/bin /usr/local/bin
