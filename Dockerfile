FROM alpine:3.5

# BCHOME is where your genesis.json, key.json and other files including state are stored.
ENV BCHOME /basecoin

# Create a basecoin user and group first so the IDs get set the same way, even
# as the rest of this may change over time.
RUN addgroup basecoin && \
    adduser -S -G basecoin basecoin

RUN mkdir -p $BCHOME && \
    chown -R basecoin:basecoin $BCHOME
WORKDIR $BCHOME

# Expose the basecoin home directory as a volume since there's mutable state in there.
VOLUME $BCHOME

# jq and curl used for extracting `pub_key` from private validator while
# deploying tendermint with Kubernetes. It is nice to have bash so the users
# could execute bash commands.
RUN apk add --no-cache bash curl jq

COPY basecoin /usr/bin/basecoin

ENTRYPOINT ["basecoin"]

# By default you will get the basecoin with local MerkleEyes and in-proc Tendermint.
CMD ["start", "--dir=${BCHOME}"]
