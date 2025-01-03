FROM golang:1.23-alpine AS build

RUN apk add build-base git linux-headers libc-dev
RUN go install github.com/go-delve/delve/cmd/dlv@latest

WORKDIR /work
COPY go.mod go.sum /work/
COPY math/go.mod math/go.sum /work/math/
COPY api/go.mod api/go.sum /work/api/
COPY core/go.mod core/go.sum /work/core/
COPY collections/go.mod collections/go.sum /work/collections/
COPY store/go.mod store/go.sum /work/store/
COPY x/tx/go.mod x/tx/go.sum /work/x/tx/
COPY x/bank/go.mod x/bank/go.sum /work/x/bank/
COPY x/staking/go.mod x/staking/go.sum /work/x/staking/
RUN go mod download

COPY ./ /work
RUN LEDGER_ENABLED=false make COSMOS_BUILD_OPTIONS="v2,debug,nostrip" clean build


FROM alpine AS run
RUN apk add bash curl jq
EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
VOLUME /simd
WORKDIR /simd

COPY contrib/images/simd-dlv/wrapper.sh /usr/bin/wrapper.sh
COPY --from=build /work/build/simdv2 /simd/
COPY --from=build /go/bin/dlv /usr/local/bin
