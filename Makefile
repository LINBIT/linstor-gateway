PROG := linstor-gateway
GITHASH=$(shell git describe --abbrev=0 --always)
GOSOURCES=$(shell find . -type f -name "*.go" -not -path "./vendor/*" -printf "%P\n")
DESTDIR =

ifndef VERSION
# default to latest git tag
VERSION=$(shell git describe --abbrev=0 --tags | tr -d 'v')
endif

all: linstor-gateway

linstor-gateway: $(GOSOURCES)
	NAME="$@"; \
	[ -n "$(GOOS)" ] && NAME="$${NAME}-$(GOOS)"; \
	[ -n "$(GOARCH)" ] && NAME="$${NAME}-$(GOARCH)"; \
	go build -o "$$NAME" \
		-ldflags "-X github.com/LINBIT/linstor-gateway/cmd.version=$(VERSION) \
		-X 'github.com/LINBIT/linstor-gateway/cmd.builddate=$(shell LC_ALL=C date --utc)' \
		-X github.com/LINBIT/linstor-gateway/cmd.githash=$(GITHASH)"

.PHONY: install
install:
	install -D -m 0750 $(PROG) $(DESTDIR)/usr/sbin/$(PROG)
	install -d -m 0750 $(DESTDIR)/etc/linstor-gateway
	install -D -m 0644 $(PROG).service $(DESTDIR)/usr/lib/systemd/system/$(PROG).service

.PHONY: release
release:
	make --always-make linstor-gateway GOOS=linux GOARCH=amd64

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
	go test ./...

.PHONY: prepare-release
prepare-release: test md-doc
	go mod tidy

vendor:
	go mod vendor

.PHONY: debrelease
debrelease: vendor checkVERSION
	dh_clean || true
	echo "VERSION=$(VERSION)" > version.env
	echo "GITHASH=$(GITHASH)" >> version.env
	tar --transform="s,^,linstor-gateway-$(VERSION)/," --owner=0 --group=0 -czf linstor-gateway-$(VERSION).tar.gz \
		$(GOSOURCES) go.mod go.sum vendor version.env \
		debian linstor-gateway.spec linstor-gateway.service linstor-gateway.xml
	rm -f version.env

ifndef VERSION
checkVERSION:
	$(error environment variable VERSION is not set)
else
checkVERSION:
ifdef FORCE
	true
else
	test -z "$$(git ls-files -m)" || $(error Uncommitted changes in working directory)
	lbvers.py check --base=$(BASE) --build=$(BUILD) --build-nr=$(BUILD_NR) --pkg-nr=$(PKG_NR) \
		--rpm-spec=linstor-gateway.spec --debian-changelog=debian/changelog --changelog=CHANGELOG.md
endif
endif


clean:
	rm -f linstor-gateway
