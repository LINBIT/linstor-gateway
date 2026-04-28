# Knobs you'll typically want to override on the command line:
#   TORUN                       comma-separated test names (default: all)
#   NVMS                        max parallel VMs (default: 12)
#   LINSTOR_GATEWAY_VERSION     yum version glob to install
#
# Knobs that pick up sensible defaults but can be overridden via env:
#   LINBIT_DOCKER_REGISTRY      docker registry for the base test-suite image
#   LINBIT_REGISTRY_URL         http(s) URL of the LINBIT package mirror
#   CI_COMMIT_SHA               short git SHA used in the default version

# Use the local checkout's HEAD when CI_COMMIT_SHA isn't provided.
CI_COMMIT_SHA ?= $(shell git -C .. rev-parse --short HEAD 2>/dev/null)

LINBIT_DOCKER_REGISTRY ?= registry.at.linbit.com
LINBIT_REGISTRY_URL    ?=

TEST_SUITE_IMAGE       ?= linstor-gateway-e2e:latest
BASE_OS_IMAGE_SOURCE   ?= https://vault.almalinux.org/9.5/cloud/x86_64/images/AlmaLinux-9-GenericCloud-9.5-20241120.x86_64.qcow2
BASE_IMAGE             ?= $(LINBIT_DOCKER_REGISTRY)/linstor-gateway-tests/base:alma-9
LINSTOR_GATEWAY_VERSION ?= 0.0.0.$(CI_COMMIT_SHA)
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
	@echo "  make base_image         build the base virter image (needs LINBIT_REGISTRY_URL)"
	@echo "  make docker             build the test-suite docker image"
	@echo "  make clean              remove build artifacts"
	@echo ""
	@echo "Knobs (override on the make command line, e.g. make run NVMS=4):"
	@echo "  TORUN                       tests to run                    [$(TORUN)]"
	@echo "  NVMS                        max parallel VMs                [$(NVMS)]"
	@echo "  LINSTOR_GATEWAY_VERSION     yum package version             [$(LINSTOR_GATEWAY_VERSION)]"
	@echo "  TEST_SUITE_IMAGE            test-suite docker image tag     [$(TEST_SUITE_IMAGE)]"
	@echo "  BASE_IMAGE                  base virter image tag           [$(BASE_IMAGE)]"
	@echo "  LINBIT_DOCKER_REGISTRY      docker registry                 [$(LINBIT_DOCKER_REGISTRY)]"
	@echo "  LINBIT_REGISTRY_URL         LINBIT package mirror URL       [$(LINBIT_REGISTRY_URL)]"

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
			--torun $(TORUN)                                                    \
			--set values.LinstorGatewayVersion=$(LINSTOR_GATEWAY_VERSION)

binary:
	$(MAKE) -C .. linstor-gateway
	mkdir -p files
	cp ../linstor-gateway files/linstor-gateway

local: binary
	$(MAKE) run LOCAL_BINARY=$(LOCAL_BINARY_CONTAINER_PATH)

base_image:
	@if [ -z "$(LINBIT_REGISTRY_URL)" ]; then \
		echo "error: LINBIT_REGISTRY_URL must be set to build the base image" >&2; \
		exit 1; \
	fi
	virter image pull linstor-gateway-tests-base-os $(BASE_OS_IMAGE_SOURCE)
	virter image build linstor-gateway-tests-base-os \
			--vcpus 4 --memory 2G \
			--provision virter/provision-base.toml \
			--set values.LinbitRepoBaseURL=$(LINBIT_REGISTRY_URL)/repository/packages-linbit-com \
			--set values.CIRepoBaseURL=$(LINBIT_REGISTRY_URL)/repository/ci-yum/rhel9 \
			$(BASE_IMAGE)

docker:
	docker build -t $(TEST_SUITE_IMAGE) .

clean:
	rm -rf files tests-out
