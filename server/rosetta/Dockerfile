FROM golang:1.15-alpine as BUILD

WORKDIR /build
COPY . ./

RUN GCO_ENABLED=0 go build -o simd ./simapp/simd/

FROM alpine

WORKDIR /app
COPY --from=BUILD /build/simd simd
RUN ./simd init test