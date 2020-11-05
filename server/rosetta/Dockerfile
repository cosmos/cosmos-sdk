FROM golang:1.15-alpine as BUILD

WORKDIR /build
COPY . ./

RUN GCO_ENABLED=0 go build -o rosetta ./cmd/rosetta

FROM alpine

WORKDIR /app
COPY --from=BUILD /build/rosetta rosetta
CMD ["sh"]
