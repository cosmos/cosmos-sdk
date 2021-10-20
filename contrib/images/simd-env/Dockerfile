FROM golang:1.17-alpine AS build
RUN apk add build-base git linux-headers
COPY ./ /work
WORKDIR /work
RUN make clean build-linux

FROM alpine:3.14 AS run
RUN apk add curl file jq
COPY contrib/images/simd-env/wrapper.sh /usr/bin/wrapper.sh

VOLUME /simd
COPY --from=build /work/build /simd/
WORKDIR /simd

EXPOSE 26656 26657
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start", "--log_format", "plain"]
STOPSIGNAL SIGTERM
