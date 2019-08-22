all: linstor-iscsi

.PHONY: linstor-iscsi
linstor-iscsi:
	GO111MODULE=on go build

.PHONY: prepare-release
prepare-release: linstor-iscsi
	# GO111MODULE=on go mod tidy
	./linstor-iscsi docs
