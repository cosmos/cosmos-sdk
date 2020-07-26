#FROM alpine:3.10.2
#
#RUN apk update && \
#    apk upgrade && \
#    apk --no-cache add curl jq file

# Changed from Alpine to Ubuntu because the keyring PR is linking to libc
# Alpine uses muslc instead of libc

FROM ubuntu:18.04

RUN apt-get update && \
    apt-get -y upgrade && \
    apt-get -y install curl jq file

VOLUME [ /simd ]
WORKDIR /simd
EXPOSE 26656 26657 1317
ENTRYPOINT ["/usr/bin/wrapper.sh"]
CMD ["start"]
STOPSIGNAL SIGTERM

COPY wrapper.sh /usr/bin/wrapper.sh
