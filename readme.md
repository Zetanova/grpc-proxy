gRPC proxy is a Go reverse proxy over UDS and/or H2C

## usage

required environment vars:
BIND_TO: sets the listening url as http or UDS address
PROXY_TO: sets the upstream proxy as http or UDS address


## docker
```
docker run --rm -d \
    -e BIND_TO=unix:///csi/csi.sock \
    -e PROXY_TO=unix:///csi/hyperv.sock \
    zetanova/grpc-poxy:1.0.1
```