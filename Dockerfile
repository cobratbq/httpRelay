FROM golang:1.21.0
COPY . /httpRelay
WORKDIR /httpRelay
RUN make all

FROM alpine:3.18.3
COPY --from=0 /httpRelay/proxy ./
COPY --from=0 /httpRelay/relay ./
