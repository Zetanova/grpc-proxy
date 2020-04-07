FROM golang:alpine as builder

RUN apk add --no-cache \
    git \
    ca-certificates

COPY ./src/ $GOPATH/src/grpc-proxy/

WORKDIR $GOPATH/src/grpc-proxy

RUN go get \
 && go build -o $GOPATH/bin

FROM alpine:3.11

COPY --from=builder /go/bin/grpc-proxy /usr/bin/grpc-proxy

CMD ["grpc-proxy"]
