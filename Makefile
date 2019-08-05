makeflags += --warn-undefined-variables

outdir = bin
tag = latest
build_tag = build
image = cobratbq/httprelay

targets = relay proxy

default: build
image-tag: image-tag-$(tag)
image-push: image-push-$(tag)

.PHONY: $(targets)
$(targets):
	CGO_ENABLED=0 go build -ldflags "-s" -installsuffix cgo -x -v -o $(outdir)/$@ ./cmd/$@

build:
	$(MAKE) relay proxy

clean:
	rm -fv $(outdir)/relay $(outdir)/proxy

docker-%:
	docker run --rm \
		-v $(PWD):/go/src/github.com/cobratbq/httpRelay \
		-v $(PWD)/image/bin:/out \
		-w /go/src/github.com/cobratbq/httpRelay \
		golang:latest \
		make $* outdir=/out

image: docker-build
	docker build -t $(image):$(build_tag) .

image-tag-%:
	docker tag -f $(image):$(build_tag) $(image):$*

image-push-%: image-tag-%
	docker push $(image):$*

