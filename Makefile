GITHASH=$(shell git describe --abbrev=0 --always)

ifndef VERSION
# default to latest git tag
VERSION=$(shell git describe --abbrev=0 --tags | tr -d 'v')
endif

all: linstor-iscsi linstor-nfs

linstor-gateway:
	GO111MODULE=on go build \
		-ldflags "-X github.com/LINBIT/linstor-gateway/cmd.version=$(VERSION) \
		-X 'github.com/LINBIT/linstor-gateway/cmd.builddate=$(shell LC_ALL=C date --utc)' \
		-X github.com/LINBIT/linstor-gateway/cmd.githash=$(GITHASH)"

linstor-iscsi linstor-nfs: linstor-gateway
	ln -s linstor-gateway $@

# internal, public doc on swagger
docs/rest/index.html: docs/rest_v1_openapi.yaml
	docker run --user="$$(id -u):$$(id -g)" --rm -v $$PWD/docs:/local \
		openapitools/openapi-generator-cli generate -i /local/rest_v1_openapi.yaml -g html -o /local/rest

# internal, public doc on swagger
api-doc: docs/rest/index.html

.PHONY: md-doc
md-doc: linstor-gateway
	./linstor-gateway docs

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: prepare-release
prepare-release: test md-doc
	GO111MODULE=on go mod tidy

linstor-gateway-$(VERSION).tar.gz: linstor-gateway
	strip linstor-gateway
	dh_clean || true
	tar --transform="s,^,linstor-gateway-$(VERSION)/," --owner=0 --group=0 -czf $@ \
		linstor-gateway debian linstor-gateway.spec linstor-iscsi.service \
		linstor-nfs.service linstor-iscsi.xml linstor-nfs.xml

debrelease: linstor-gateway-$(VERSION).tar.gz

clean:
	rm linstor-gateway linstor-iscsi linstor-nfs
