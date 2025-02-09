FROM golang:1.22-alpine3.21
COPY . /httpRelay
WORKDIR /httpRelay
RUN go build -tags netgo ./cmd/proxy
RUN go build -tags netgo ./cmd/relay

FROM alpine:3.21
COPY --from=0 /httpRelay/proxy ./
COPY --from=0 /httpRelay/relay ./

