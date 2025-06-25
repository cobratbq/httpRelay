FROM golang:1.23-alpine3.22
COPY . /httpRelay
WORKDIR /httpRelay
RUN go build -buildmode pie -tags netgo ./cmd/proxy
RUN go build -buildmode pie -tags netgo ./cmd/relay

FROM alpine:3.22
COPY --from=0 /httpRelay/proxy ./
COPY --from=0 /httpRelay/relay ./

