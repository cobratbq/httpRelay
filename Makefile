# Applications are built by default using 'netgo' build tag.
#
# Build for other architectures: `GOARCH=arm make`

.PHONY: all
all: library proxy relay

.PHONY: library
library:
	go build -tags netgo ./...

.PHONY: proxy
proxy: library
	go build -tags netgo ./cmd/proxy

.PHONY: relay
relay: library
	go build -tags netgo ./cmd/relay

.PHONY: test
test:
	go test -tags netgo

.PHONY: clean
clean:
	rm -f proxy relay
