FROM gliderlabs/alpine
RUN apk add --no-cache ca-certificates

COPY image/entrypoint /
ENTRYPOINT ["/entrypoint"]

COPY image/bin/relay image/bin/proxy /usr/local/bin/
USER nobody

