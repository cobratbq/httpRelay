FROM golang:1.21-alpine3.18
COPY . /httpRelay
WORKDIR /httpRelay
RUN go build -tags netgo ./cmd/proxy
RUN go build -tags netgo ./cmd/relay

FROM alpine:3.18
COPY --from=0 /httpRelay/proxy ./
COPY --from=0 /httpRelay/relay ./

