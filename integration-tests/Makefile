# Knobs you'll typically want to override on the command line:
#   TORUN                       comma-separated test names (default: all)
#   NVMS                        max parallel VMs (default: 12)
#   LINSTOR_GATEWAY_VERSION     yum version glob to install
#   DRBD_VERSION                drbd version glob (empty = latest stable from drbd repo)
#   DRBD_UTILS_VERSION          drbd-utils yum version glob
#   LINSTOR_VERSION             linstor-{controller,satellite,client} yum version glob
#   DRBD_REACTOR_VERSION        drbd-reactor yum version glob
#   NFS_GANESHA_SOURCE          "linbit" (drbd-9 repo) or "ci" (pulp CI repo)
#   NFS_GANESHA_VERSION         nfs-ganesha VERSION-RELEASE (required for "ci")
#
# By default these defer to the stable LINBIT repo (latest released versions)
# so a broken upstream master doesn't block this project's MRs. To target
# development builds (a future nightly job), override with `.latest` globs and
# enable the linbit-ci repo via `dnf config-manager --set-enabled linbit-ci`
# in a custom provisioning step.
#
# Knobs that pick up sensible defaults but can be overridden via env:
#   LINBIT_DOCKER_REGISTRY      docker registry for the base test-suite image
#   LINBIT_PACKAGES_URL         http(s) URL of the LINBIT package mirror
#   LINBIT_PULP_BASE_URL        http(s) URL of the LINBIT pulp server
#   CI_COMMIT_SHA               short git SHA used in the default version

# Use the local checkout's HEAD when CI_COMMIT_SHA isn't provided.
CI_COMMIT_SHA ?= $(shell git -C .. rev-parse --short HEAD 2>/dev/null)

LINBIT_DOCKER_REGISTRY ?= registry.at.linbit.com
LINBIT_PACKAGES_URL    ?= https://pkg-linbit.cache.at.linbit.com
LINBIT_PULP_BASE_URL   ?= https://pulp.at.linbit.com

TEST_SUITE_IMAGE       ?= linstor-gateway-e2e:latest
BASE_OS_IMAGE_SOURCE   ?= https://vault.almalinux.org/9.5/cloud/x86_64/images/AlmaLinux-9-GenericCloud-9.5-20241120.x86_64.qcow2
BASE_IMAGE             ?= $(LINBIT_DOCKER_REGISTRY)/linstor-gateway-tests/base:alma-9
LINSTOR_GATEWAY_VERSION ?= 0.0.0.$(CI_COMMIT_SHA)
# Empty defaults = latest released version from the stable drbd repo.
# Override on the command line (or in a nightly job) to pin a specific version
# or to target latest CI builds (e.g. 9.3.0.latest, 1.99.0.latest-*).
DRBD_VERSION           ?=
DRBD_UTILS_VERSION     ?=
LINSTOR_VERSION        ?=
DRBD_REACTOR_VERSION   ?=
# The CI repo holds many builds, so NFS_GANESHA_SOURCE=ci needs a pinned
# NFS_GANESHA_VERSION; for "linbit", empty means the latest release.
NFS_GANESHA_SOURCE     ?= linbit
NFS_GANESHA_VERSION    ?=
NFS_GANESHA_CI_REPO_URL ?= $(LINBIT_PULP_BASE_URL)/pulp/content/rpm/rhel9/
TORUN                  ?= all
NVMS                   ?= 12

# When set, the testimage provisioning step copies this file onto the VMs
# instead of installing the linstor-gateway yum package. The path is
# evaluated inside the ansible-runner container, where vmshed mounts
# integration-tests/ at /virter/workspace.
LOCAL_BINARY                 ?=
LOCAL_BINARY_CONTAINER_PATH  := /virter/workspace/files/linstor-gateway

.PHONY: help run local binary base_image docker clean

help:
	@echo "Common workflows:"
	@echo "  make run [TORUN=...]    run tests against the linstor-gateway yum package"
	@echo "  make local [TORUN=...]  build the gateway from this checkout and run against it"
	@echo ""
	@echo "Other targets:"
	@echo "  make binary             just build the gateway into ./files/linstor-gateway"
	@echo "  make base_image         build the base virter image"
	@echo "  make docker             build the test-suite docker image"
	@echo "  make clean              remove build artifacts"
	@echo ""
	@echo "Knobs (override on the make command line, e.g. make run NVMS=4):"
	@echo "  TORUN                       tests to run                    [$(TORUN)]"
	@echo "  NVMS                        max parallel VMs                [$(NVMS)]"
	@echo "  LINSTOR_GATEWAY_VERSION     yum package version             [$(LINSTOR_GATEWAY_VERSION)]"
	@echo "  DRBD_VERSION                drbd yum version glob           [$(DRBD_VERSION)]"
	@echo "  DRBD_UTILS_VERSION          drbd-utils yum version          [$(DRBD_UTILS_VERSION)]"
	@echo "  LINSTOR_VERSION             linstor-* yum version           [$(LINSTOR_VERSION)]"
	@echo "  DRBD_REACTOR_VERSION        drbd-reactor yum version        [$(DRBD_REACTOR_VERSION)]"
	@echo "  NFS_GANESHA_SOURCE          nfs-ganesha source (linbit|ci)  [$(NFS_GANESHA_SOURCE)]"
	@echo "  NFS_GANESHA_VERSION         nfs-ganesha VERSION-RELEASE     [$(NFS_GANESHA_VERSION)]"
	@echo "  NFS_GANESHA_CI_REPO_URL     nfs-ganesha CI yum repo         [$(NFS_GANESHA_CI_REPO_URL)]"
	@echo "  TEST_SUITE_IMAGE            test-suite docker image tag     [$(TEST_SUITE_IMAGE)]"
	@echo "  BASE_IMAGE                  base virter image tag           [$(BASE_IMAGE)]"
	@echo "  LINBIT_DOCKER_REGISTRY      docker registry                 [$(LINBIT_DOCKER_REGISTRY)]"
	@echo "  LINBIT_PACKAGES_URL         LINBIT package mirror URL       [$(LINBIT_PACKAGES_URL)]"
	@echo "  LINBIT_PULP_BASE_URL        LINBIT pulp server URL          [$(LINBIT_PULP_BASE_URL)]"

run:
	mkdir -p tests-out
	vmshed																		\
			--quiet																\
			--out-dir "$(shell readlink -f tests-out)"                          \
			--startvm 20                                                        \
			--nvms $(NVMS)                                                      \
			--vms virter/vms.toml                                               \
			--tests virter/tests.toml                                           \
			--set values.TestSuiteImage=$(TEST_SUITE_IMAGE)                     \
			--set values.LinstorGatewayLocalBinary=$(LOCAL_BINARY)              \
			--set values.DrbdVersion=$(DRBD_VERSION)                            \
			--set values.DrbdUtilsVersion=$(DRBD_UTILS_VERSION)                 \
			--set values.LinstorVersion=$(LINSTOR_VERSION)                      \
			--set values.DrbdReactorVersion=$(DRBD_REACTOR_VERSION)             \
			--set values.NfsGaneshaSource=$(NFS_GANESHA_SOURCE)                 \
			--set values.NfsGaneshaVersion=$(NFS_GANESHA_VERSION)               \
			--set values.NfsGaneshaCIRepoURL=$(NFS_GANESHA_CI_REPO_URL)         \
			--torun $(TORUN)                                                    \
			--set values.LinstorGatewayVersion=$(LINSTOR_GATEWAY_VERSION)

binary:
	$(MAKE) -C .. linstor-gateway
	mkdir -p files
	cp ../linstor-gateway files/linstor-gateway

local: binary
	$(MAKE) run LOCAL_BINARY=$(LOCAL_BINARY_CONTAINER_PATH)

base_image:
	virter image pull linstor-gateway-tests-base-os $(BASE_OS_IMAGE_SOURCE)
	virter image build linstor-gateway-tests-base-os \
			--vcpus 4 --memory 2G \
			--provision virter/provision-base.toml \
			--set values.LinbitRepoBaseURL=$(LINBIT_PACKAGES_URL) \
			--set values.CIRepoBaseURL=$(LINBIT_PULP_BASE_URL)/pulp/content/rpm/rhel9/ \
			$(BASE_IMAGE)

docker:
	docker build -t $(TEST_SUITE_IMAGE) .

clean:
	rm -rf files tests-out
