#!/bin/bash

set -e

die() {
	echo >&2 "$1"
	exit 1
}

version=$1
[ -z "$version" ] && die "Usage: $0 <version>"

export EMAIL="$(git config --get user.email)"
export NAME="$(git config --get user.name)"

version_and_release="${version}-1"

rpmdev-bumpspec -n "$version_and_release" \
	-c "New upstream release" \
	-u "$NAME <$EMAIL>" \
	linstor-gateway.spec

dch -v "$version_and_release" \
	-u "medium" \
	"New upstream release" \
	&& dch -r ""

date=$(date -u +"%Y-%m-%d")
sed -i '9r /dev/stdin' CHANGELOG.md <<EOF
## $version - $date

### Changes

TODO

### Fixes

TODO

EOF

${EDITOR:-vi} CHANGELOG.md

git --no-pager diff -- CHANGELOG.md linstor-gateway.spec debian/changelog
