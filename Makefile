LATESTTAG=$(shell git describe --abbrev=0 --tags | tr -d 'v')

all: linstor-iscsi

.PHONY: linstor-iscsi
linstor-iscsi:
	GO111MODULE=on go build

# internal, public doc on swagger
docs/rest/index.html: docs/rest_v1_openapi.yaml
	docker run --user="$$(id -u):$$(id -g)" --rm -v $$PWD/docs:/local \
		openapitools/openapi-generator-cli generate -i /local/rest_v1_openapi.yaml -g html -o /local/rest

# internal, public doc on swagger
api-doc: docs/rest/index.html

.PHONY: md-doc
md-doc: linstor-iscsi
	./linstor-iscsi docs

.PHONY: test
test:
	GO111MODULE=on go test ./...

.PHONY: prepare-release
prepare-release: test md-doc
	GO111MODULE=on go mod tidy

linstor-iscsi-$(LATESTTAG).tar.gz: linstor-iscsi
	strip linstor-iscsi
	dh_clean || true
	tar --transform="s,^,linstor-iscsi-$(LATESTTAG)/," --owner=0 --group=0 -czf $@ linstor-iscsi debian linstor-iscsi.spec

debrelease: linstor-iscsi-$(LATESTTAG).tar.gz
