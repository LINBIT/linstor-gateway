TEST_SUITE_IMAGE ?= linstor-gateway-e2e:latest
BASE_IMAGE ?= linstor-gateway-tests/base:centos-8
LINSTOR_GATEWAY_VERSION ?= 0.0.0.$$CI_COMMIT_SHA

run:
	mkdir -p tests-out
	vmshed																		\
            --out-dir "$(shell readlink -f tests-out)"                          \
            --startvm 20                                                        \
            --nvms 12                                                           \
            --vms virter/vms.toml                                               \
            --tests virter/tests.toml                                           \
            --set values.TestSuiteImage=$(TEST_SUITE_IMAGE) \
            --set values.LinstorGatewayVersion=$(LINSTOR_GATEWAY_VERSION)

base_image:
	virter image build centos-8 \
			--provision virter/provision-base.toml \
			--set values.DRBDRepoBaseURL=$$LINBIT_REGISTRY_URL/repository/packages-linbit-com/yum/rhel8/drbd-9/x86_64 \
			--set values.PacemakerRepoBaseURL=$$LINBIT_REGISTRY_URL/repository/packages-linbit-com/yum/rhel8/pacemaker-2/x86_64 \
			--set values.CIRepoBaseURL=$$LINBIT_REGISTRY_URL/repository/rhel8 \
			$(BASE_IMAGE)