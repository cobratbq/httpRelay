# Applications are built by default using 'netgo' build tag.
#
# Build for other architectures: `GOARCH=arm make`

.PHONY: all
all: library proxy relay

.PHONY: library
library:
	go build ./...

.PHONY: proxy
proxy: library
	go build -buildmode pie ./cmd/proxy

.PHONY: relay
relay: library
	go build -buildmode pie ./cmd/relay

.PHONY: test
test: library
	go test

.PHONY: docker
docker:
	docker build --security-opt label=disable -t httprelay .

.PHONY: podman
podman:
	podman build --security-opt label=disable -t httprelay .

.PHONY: clean
clean:
	rm -f proxy relay
