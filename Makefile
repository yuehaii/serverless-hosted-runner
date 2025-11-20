SHELL=/bin/bash
-include ./env.sh
include ./Registration.mk
-include ./registration/Registration_$(ralias).mk

local_image_v := v0.1.1
cur_dir := $(shell pwd)
table_addr := cn-shanghai.ots.aliyuncs.com
az_storage_account := slsrunnertfstate
oss_alias := $(shell echo $(SLS_CD_ENV) | grep Prod > /dev/null && echo $(ralias) || echo "test")
tf_alias := $(ralias)
SLS_CLOUD_PR := $(shell [ -z $(SLS_CLOUD_PR) ] && echo ali || echo $(SLS_CLOUD_PR))
table_alias := $(shell echo $(SLS_CD_ENV) | grep Prod > /dev/null && echo $(ralias) || echo "test")
image_v := $(shell [ -z $(BUILD_IMAGE_VER) ] && echo $(BUILD_IMAGE_RUN_NUM_VER) || echo $(BUILD_IMAGE_VER))
repo_image := $(shell [ -z $(SLS_REPO_IMAGE) ] && echo runner || echo $(SLS_REPO_IMAGE))
repo_image_ver := $(shell [ -z $(SLS_REPO_IMAGE_VERSION) ] && echo 14 || echo $(SLS_REPO_IMAGE_VERSION))
# duser := $(shell [ -z $(SLS_CLOUD_PR) ] || [ $(SLS_CLOUD_PR) == "ali" ] && echo runner || echo root)
duser := root # for mix cloud
dispacher_cpu := $(shell ([ $(SLS_CLOUD_PR) == "ali" ] && echo $(WF_SLS_DISPACHER_CPU_SIZE_ALI)) || ([ $(SLS_CLOUD_PR) == "azure" ] && echo $(WF_SLS_DISPACHER_CPU_SIZE_AZURE)) || ([ $(SLS_CLOUD_PR) == "gcp" ] && echo $(WF_SLS_DISPACHER_CPU_SIZE_GCP)) || echo $(WF_SLS_DISPACHER_CPU_SIZE))
dispacher_memory := $(shell ([ $(SLS_CLOUD_PR) == "ali" ] && echo $(WF_SLS_DISPACHER_MEM_SIZE_ALI)) || ([ $(SLS_CLOUD_PR) == "azure" ] && echo $(WF_SLS_DISPACHER_MEM_SIZE_AZURE)) || ([ $(SLS_CLOUD_PR) == "gcp" ] && echo $(WF_SLS_DISPACHER_MEM_SIZE_GCP)) || echo $(WF_SLS_DISPACHER_MEM_SIZE))

.PHONY: all lazy_install mix_install lazy_install_kube agent_install clean debugger_install image_remoted lazy_install_remoted gen_pb kafka_certs push_dispacher push_runner lazy_destroy
all: image install test
image: gen_certs dipatcher_image runner_image clean_certs
install: allen_install local_install

lazy_destroy:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); cld_pr_state=$(SLS_CLOUD_PR_STATE); \
	[[ "$$cld_pr" = "ali" && "$$cld_pr_state" = "billing" ]] && { terraform init; }; \
	[[ "$$cld_pr" = "ali" ]] && { terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ "$$cld_pr" = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	[[ "$$cld_pr" = "gcp" ]] && { terraform init \
		-backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="prefix=terraform/state/$(tf_alias)"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); }; \
	echo "DISPATCHER_VSWITCH_ID is $(DISPATCHER_VSWITCH_ID)"; echo "ori_vs_id is $$ori_vs_id"; \
	echo "dispacher_cpu is $(dispacher_cpu)"; echo "dispacher_memory is $(dispacher_memory)"; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	terraform destroy -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; destroy_code=$$?; echo "destroy_code is: $$destroy_code"; \
	[[ $$destroy_code -ne 0 ]] && terraform destroy -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="slb_id=$$slb_id" -var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; \
	exit 0

lazy_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); cld_pr_state=$(SLS_CLOUD_PR_STATE); \
	[[ "$$cld_pr" = "ali" && "$$cld_pr_state" = "billing" ]] && { terraform init; }; \
	[[ "$$cld_pr" = "ali" ]] && { terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ "$$cld_pr" = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	[[ "$$cld_pr" = "gcp" ]] && { terraform init \
		-backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="prefix=terraform/state/$(tf_alias)"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); }; \
	echo "DISPATCHER_VSWITCH_ID is $(DISPATCHER_VSWITCH_ID)"; echo "ori_vs_id is $$ori_vs_id"; \
	echo "dispacher_cpu is $(dispacher_cpu)"; echo "dispacher_memory is $(dispacher_memory)"; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="slb_id=$$slb_id" -var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; \
	exit 0

allen_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); \
	[[ "$$cld_pr" = "ali" ]] && { terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ "$$cld_pr" = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); slb_id=$(DISPATCHER_SLB_ID); }; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	echo "dispacher_cpu is $(dispacher_cpu)"; echo "dispacher_memory is $(dispacher_memory)"; \
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="slb_id=$$slb_id" -auto-approve; \
	exit 0

mix_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); \
	[[ "$$cld_pr" = "ali" ]] && { echo "ali pr"; terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ "$$cld_pr" = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init -upgrade; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); }; \
	echo "DISPATCHER_VSWITCH_ID is $(DISPATCHER_VSWITCH_ID)"; echo "ori_vs_id is $$ori_vs_id"; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	echo "subnet_ids is $(AZ_SUBNET_IDS)"; echo "resource_group_name is $(AZ_RG_NAME)"; \
	echo "dispacher_cpu is $(dispacher_cpu)"; echo "dispacher_memory is $(dispacher_memory)"; \
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" -var="tf_ctl=$(SLS_TF_CTL)" \
		-var="dispacher_cpu=$(dispacher_cpu)" -var="dispacher_memory=$(dispacher_memory)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="slb_id=$$slb_id" -var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; \
	exit 0

lazy_install_kube: 
	kubectl create deployment sls-dispacher --image=artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) -- ./dispatcher -v $(image_v) -r $(WF_SLS_REGS) -n serverless-runner-dispatcher; \
	exit 0

runner_image:
	sudo echo "duser is $(duser)"; \
	docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		-f ./image/Dockerfile.$(repo_image) --build-arg SHR_C_T=runner \
		--secret id=tk-un,env=GIT_ACCESS_TOKEN_USR --secret id=tk-pw,env=GIT_ACCESS_TOKEN_PWD \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) \
		--build-arg GO_ARCH=amd64 --build-arg SLS_TF_CTL="$(SLS_TF_CTL)" \
		--build-arg $(repo_image)=$(repo_image_ver) \
		--build-arg RUNNER_PLATFORM=linux-x64 --build-arg DOCKER_USER=$(duser) .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	[[ -z $$GOOGLE_CREDENTIALS ]] || { echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://$(GOOGLE_REGION)-docker.pkg.dev; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker tag $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:latest; \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:latest; }
	[[ -z $$AZURE_ACR_USRNAME ]] || { docker login -u "$$AZURE_ACR_USRNAME" -p "$$AZURE_ACR_PWD" "$$AZURE_ACR_SERVER"; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:latest; \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:latest; }

dipatcher_image:
	sudo echo "building dispacher service image ..."; \
	docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) \
		--build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=amd64 \
		--build-arg GO_ARCH=amd64 --build-arg SLS_TF_CTL="$(SLS_TF_CTL)" \
		--secret id=tk-un,env=GIT_ACCESS_TOKEN_USR --secret id=tk-pw,env=GIT_ACCESS_TOKEN_PWD \
		--secret id=kafka-ca,env=KAFKA_INS_CA_CERT .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	[[ -z $$GOOGLE_CREDENTIALS ]] || { echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://$(GOOGLE_REGION)-docker.pkg.dev; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker tag $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:latest; \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:latest; }
	[[ -z $$AZURE_ACR_USRNAME ]] || { docker login -u "$$AZURE_ACR_USRNAME" -p "$$AZURE_ACR_PWD" "$$AZURE_ACR_SERVER"; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:latest; \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:latest; }

push_dispacher:
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	[[ -z $$GOOGLE_CREDENTIALS ]] || { echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://$(GOOGLE_REGION)-docker.pkg.dev; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker tag $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:latest; \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-dispatcher:latest; }
	[[ -z $$AZURE_ACR_USRNAME ]] || { docker login -u "$$AZURE_ACR_USRNAME" -p "$$AZURE_ACR_PWD" "$$AZURE_ACR_SERVER"; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:$(image_v); \
    docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:latest; \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-dispatcher:latest; }

push_runner:
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	[[ -z $$GOOGLE_CREDENTIALS ]] || { echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://$(GOOGLE_REGION)-docker.pkg.dev; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker tag $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v) \
        $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:latest; \
    docker push $(GOOGLE_REGION)-docker.pkg.dev/$(GOOGLE_PROJECT)/serverless-hosted-runner/serverless-hosted-runner-eci:latest; }
	[[ -z $$AZURE_ACR_USRNAME ]] || { docker login -u "$$AZURE_ACR_USRNAME" -p "$$AZURE_ACR_PWD" "$$AZURE_ACR_SERVER"; \
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:$(image_v); \
    docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
        $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:latest; \
    docker push $(AZURE_ACR_SERVER)/ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci:latest; }



gen_pb:
	pushd .; cd $(cur_dir)/src/network; \
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative  grpc/listener.proto; popd

gen_certs:
ifeq (${DISPACHER_CA_CERT},)
	mkdir ./src/certs; cp ./template/cert/cert.tpl ./src/certs/main.go; \
	pushd .; cd ./src; go build -o ./certs/certagent ./certs/main.go; ./certs/certagent; \
	echo "please save ca cert hash: "; awk '{printf "%s\\n", $$0}' ./certs/ca.cert.pem | base64; \
	echo " please save ca key hash: "; awk '{printf "%s\\n", $$0}' ./certs/ca.key.pem | base64; \
	popd; exit 0;
else ifeq (${DISPACHER_CA_CERT},"")
	mkdir ./src/certs; cp ./template/cert/cert.tpl ./src/certs/main.go; \
	pushd .; cd ./src; go build -o ./certs/certagent ./certs/main.go; ./certs/certagent; \
	echo "please save ca cert hash: "; awk '{printf "%s\\n", $$0}' ./certs/ca.cert.pem | base64; \
	echo " please save ca key hash: "; awk '{printf "%s\\n", $$0}' ./certs/ca.key.pem | base64; \
	popd; exit 0;
else
	mkdir ./src/certs; \
	echo -e "$(shell echo ${DISPACHER_CA_CERT} | base64 -d)" > ./src/certs/ca.cert.pem; \
	echo -e "$(shell echo ${DISPACHER_CA_KEY} | base64 -d)" > ./src/certs/ca.key.pem; exit 0;
endif

kafka_cert:
	echo -e "$(shell cat ${certfile} | base64)"

clean_certs:
	rm -rf ./src/certs; exit 0

agent_install:
	cd $(cur_dir)/agent/$(SLS_CLOUD_PR); export ALICLOUD_REGION=$(AGENT_REGION); \
	terraform init; init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	terraform apply -var-file="ubuntu_agent.tfvars" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var-file="ubuntu_agent.tfvars" -auto-approve; \
	exit 0  

local_install:
	docker run --name local-dispacher --env distestmode=local \
		-it artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		./dispatcher -v $(image_v) -r $(WF_SLS_REGS) -a none -m 1 -c $(SLS_CLOUD_PR) -t $(SLS_TF_CTL); \
	exit 0

clean:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); \
	terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; \
    terraform destroy -var-file="ubuntu_dispatcher.tfvars" -auto-approve; \
    cd $(cur_dir)/agent/$(SLS_CLOUD_PR); export ALICLOUD_REGION=$(AGENT_REGION); \
	terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; \
    terraform destroy -var-file="ubuntu_agent.tfvars" -auto-approve; cd $(cur_dir)

.PHONY: local local_clean

local: local_image local_istall

local_istall: 
	sudo docker-compose -f ./image/docker-compose.yaml build
	sudo docker-compose -f ./image/docker-compose.yaml up -d

local_image: local_dispatcher_image local_runner_image

local_dispatcher_image:
	sudo docker build \
		-t localhost/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(local_image_v) \
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=arm64 \
		--build-arg GO_ARCH=amd64 --build-arg LOCAL_MODE=True .  
		
local_runner_image:
	sudo docker build \
		-t localhost/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(local_image_v) \
		-f ./image/Dockerfile.runner --build-arg SHR_C_T=runner \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) \
		--build-arg GO_ARCH=amd64 --build-arg RUNNER_PLATFORM=linux-x64 . 

local_clean:
	sudo docker-compose -f ./image/docker-compose.yaml down
	sudo docker-compose -f ./image/docker-compose.yaml rm
	sudo docker rmi localhost/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(local_image_v)
	sudo docker rmi localhost/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(local_image_v)

debugger_install:
	sudo docker buildx build --platform linux/amd64 \
			-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/allen-db-debugger:latest \
			-f ./tool/test/Dockerfile.debuger .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/allen-db-debugger:latest
	kubectl create secret docker-registry image-pull-secret \
			--docker-server=artifactory.cloud.ingka-system.cn \
			--docker-username=$(DOCKER_USERNAME) \
			--docker-password=$(DOCKER_PWD) -n debugger --kubeconfig ~/.kube/config
	kubectl apply -f ./tool/test/debuger.yaml -n debugger --kubeconfig ~/.kube/config

debugger_nprod_install:
	sudo docker buildx build --platform linux/amd64 \
			-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/allen-db-debugger-nprod:latest \
			-f ./tool/ccoecn/Dockerfile_nprod.debuger .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/allen-db-debugger-nprod:latest
	kubectl create secret docker-registry image-pull-secret \
			--docker-server=artifactory.cloud.ingka-system.cn \
			--docker-username=$(DOCKER_USERNAME) \
			--docker-password=$(DOCKER_PWD) -n panda-dev --kubeconfig ~/.kube/config_mongodb.internal
	kubectl apply -f ./tool/ccoecn/debuger_nprod.yaml -n panda-dev --kubeconfig ~/.kube/config_mongodb.internal

test:
	echo $(SLS_CLOUD_PR)
	exit 0

### deprecated self-created remote dockerd entries. 
### the self-created dockerd server already been removed.
# ali already support the dind in tf pr v1.225.1 
# https://github.com/aliyun/terraform-provider-alicloud/releases/tag/v1.225.1
image_remoted: dipatcher_image_remoted runner_image_remoted

lazy_install_remoted: 
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); \
	terraform init; init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); }; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	terraform apply -var="image_ver=$(image_v)-rd" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)-rd" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="slb_id=$$slb_id" -var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; \
	exit 0
runner_image_remoted:
	sudo docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)-rd \
		-f ./image/Dockerfile.runner.remoted --build-arg SHR_C_T=runner \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) \
		--build-arg GO_ARCH=amd64 \
		--build-arg RUNNER_PLATFORM=linux-x64 .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)-rd
dipatcher_image_remoted:
	sudo docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)-rd \
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) \
		--build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=amd64 \
		--build-arg GO_ARCH=amd64 .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)-rd