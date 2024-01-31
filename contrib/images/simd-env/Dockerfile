FROM golang:1.19-alpine AS build

RUN apk add build-base git linux-headers

WORKDIR /work
COPY go.mod go.sum /work/
COPY db/go.mod db/go.sum /work/db/
COPY errors/go.mod errors/go.sum /work/errors/
COPY math/go.mod math/go.sum /work/math/

RUN go mod download
COPY ./ /work
RUN LEDGER_ENABLED=false make clean build

FROM alpine AS run
RUN apk add bash curl jq
COPY contrib/images/simd-env/wrapper.sh /usr/bin/wrapper.sh

VOLUME /simd
COPY --from=build /work/build/simd /simd/
WORKDIR /simd

EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
