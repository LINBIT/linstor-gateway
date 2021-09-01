GITHASH=$(shell git describe --abbrev=0 --always)
GOSOURCES=$(shell find . -type f -name '*.go')

ifndef VERSION
# default to latest git tag
VERSION=$(shell git describe --abbrev=0 --tags | tr -d 'v')
endif

all: linstor-gateway

linstor-gateway: $(GOSOURCES)
	NAME="$@"; \
	[ -n "$(GOOS)" ] && NAME="$${NAME}-$(GOOS)"; \
	[ -n "$(GOARCH)" ] && NAME="$${NAME}-$(GOARCH)"; \
	go build \
		-ldflags "-X github.com/LINBIT/linstor-gateway/cmd.version=$(VERSION) \
		-X 'github.com/LINBIT/linstor-gateway/cmd.builddate=$(shell LC_ALL=C date --utc)' \
		-X github.com/LINBIT/linstor-gateway/cmd.githash=$(GITHASH)"

.PHONY: release
release:
	make linstor-gateway GOOS=linux GOARCH=amd64

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

.PHONY: debrelease
debrelease: checkVERSION
	strip linstor-gateway
	dh_clean || true
	tar --transform="s,^,linstor-gateway-$(VERSION)/," --owner=0 --group=0 -czf linstor-gateway-$(VERSION).tar.gz \
		linstor-gateway debian linstor-gateway.spec linstor-gateway.service \
		linstor-gateway.xml

ifndef VERSION
checkVERSION:
	$(error environment variable VERSION is not set)
else
checkVERSION:
ifdef FORCE
	true
else
	test -z "$$(git ls-files -m)"
	lbvers.py check --base=$(BASE) --build=$(BUILD) --build-nr=$(BUILD_NR) --pkg-nr=$(PKG_NR) \
		--rpm-spec=linstor-gateway.spec --debian-changelog=debian/changelog --changelog=CHANGELOG.md
endif
endif


clean:
	rm linstor-gateway
