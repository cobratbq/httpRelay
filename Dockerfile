FROM golang:1.21.0
COPY . /httpRelay
WORKDIR /httpRelay
RUN make all

FROM golang:1.21.0-alpine
COPY --from=0 /httpRelay/proxy ./
COPY --from=0 /httpRelay/relay ./
