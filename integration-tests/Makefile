TEST_SUITE_IMAGE ?= linstor-gateway-e2e:latest
BASE_OS_IMAGE_SOURCE ?= https://vault.almalinux.org/9.5/cloud/x86_64/images/AlmaLinux-9-GenericCloud-9.5-20241120.x86_64.qcow2
BASE_IMAGE ?= $$LINBIT_DOCKER_REGISTRY/linstor-gateway-tests/base:alma-9
LINSTOR_GATEWAY_VERSION ?= 0.0.0.$$CI_COMMIT_SHA
TORUN ?= all
NVMS ?= 12

run:
	mkdir -p tests-out
	vmshed																		\
			--quiet																\
			--out-dir "$(shell readlink -f tests-out)"                          \
			--startvm 20                                                        \
			--nvms $(NVMS)                                                      \
			--vms virter/vms.toml                                               \
			--tests virter/tests.toml                                           \
			--set values.TestSuiteImage=$(TEST_SUITE_IMAGE) 					\
			--torun $(TORUN) \
			--set values.LinstorGatewayVersion=$(LINSTOR_GATEWAY_VERSION)

base_image:
	virter image pull linstor-gateway-tests-base-os $(BASE_OS_IMAGE_SOURCE)
	virter image build linstor-gateway-tests-base-os \
			--provision virter/provision-base.toml \
			--set values.LinbitRepoBaseURL=$$LINBIT_REGISTRY_URL/repository/packages-linbit-com \
			--set values.CIRepoBaseURL=$$LINBIT_REGISTRY_URL/repository/ci-yum/rhel9 \
			$(BASE_IMAGE)

docker:
	docker build -t $(TEST_SUITE_IMAGE) .
