FROM golang:1.15-alpine as BUILD

WORKDIR /build
COPY . ./

RUN go build -o simd ./simapp/simd/

FROM alpine

WORKDIR /app
COPY --from=BUILD /build/simd simd
# required otherwise simd will cry outloud
# because it cannot intercept tendermint configs
RUN ./simd init test