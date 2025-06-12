PROG := linstor-gateway

GOSOURCES=$(shell find . -type f -name "*.go" -not -path "./vendor/*" -printf "%P\n")
DESTDIR =

all: linstor-gateway

linstor-gateway: $(GOSOURCES) version.env
	. ./version.env; \
	NAME="$@"; \
	[ -n "$(GOOS)" ] && NAME="$${NAME}-$(GOOS)"; \
	[ -n "$(GOARCH)" ] && NAME="$${NAME}-$(GOARCH)"; \
	go build -o "$$NAME" \
		-ldflags "-X github.com/LINBIT/linstor-gateway/pkg/version.Version=$${VERSION} \
		-X 'github.com/LINBIT/linstor-gateway/pkg/version.BuildDate=$(shell LC_ALL=C date --utc)' \
		-X github.com/LINBIT/linstor-gateway/pkg/version.GitCommit=$${GITHASH}"

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

vendor: go.mod go.sum
	go mod vendor

version.env:
	if [ -n "$(VERSION)" ]; then \
		VERSION="$(VERSION)"; \
	else \
		# default to latest git tag \
		VERSION=$(shell git describe --abbrev=0 --tags | tr -d 'v'); \
	fi; \
	echo "VERSION=$${VERSION}" > version.env
	echo "GITHASH=$(shell git describe --abbrev=0 --always)" >> version.env

.PHONY: debrelease
debrelease: clean clean-version vendor version.env checkVERSION
	dh_clean || true
	tar --transform="s,^,linstor-gateway-$(VERSION)/," --owner=0 --group=0 -czf linstor-gateway-$(VERSION).tar.gz \
		$(GOSOURCES) go.mod go.sum vendor version.env Makefile \
		debian linstor-gateway.spec linstor-gateway.service linstor-gateway.xml

ifndef VERSION
checkVERSION:
	$(error environment variable VERSION is not set)
else
checkVERSION:
ifdef FORCE
	true
else
	test -z "$$(git ls-files -m)" || { echo "Uncommitted files in working directory"; exit 1; }
	lbvers.py check --base=$(BASE) --build=$(BUILD) --build-nr=$(BUILD_NR) --pkg-nr=$(PKG_NR) \
		--rpm-spec=linstor-gateway.spec --debian-changelog=debian/changelog --changelog=CHANGELOG.md
endif
endif

.PHONY: clean
clean:
	rm -f linstor-gateway

.PHONY: clean-version
clean-version:
	rm -f version.env

sbom/linstor-gateway.cdx.json:
	@mkdir -p sbom
	go run github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest app -json -output $@
