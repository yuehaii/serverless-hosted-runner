SHELL=/bin/bash
-include ./env.sh
include ./Registration.mk
-include ./registration/Registration_$(ralias).mk

image_v_latest := 1.6.1
local_image_v := $(image_v_latest)
cur_dir := $(shell pwd)
table_addr := cn-shanghai.ots.aliyuncs.com
az_storage_account := slsrunnertfstate
oss_alias := $(shell echo $(SLS_CD_ENV) | grep Prod > /dev/null && echo $(ralias) || echo "test")
tf_alias := $(ralias)
table_alias := $(shell echo $(SLS_CD_ENV) | grep Prod > /dev/null && echo $(ralias) || echo "test")
image_v := $(shell [ -z $(BUILD_IMAGE_VER) ] && echo $(image_v_latest) || echo $(BUILD_IMAGE_VER))
duser := $(shell [ -z $(SLS_CLOUD_PR) ] || [ $(SLS_CLOUD_PR) == "ali" ] && echo runner || echo root)

.PHONY: all lazy_install mix_install lazy_install_kube agent_install clean debugger_install image_remoted lazy_install_remoted
all: image install test
image: dipatcher_image runner_image
install: allen_install

lazy_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); \
	[[ $$cld_pr = "ali" ]] && { terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ $$cld_pr = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	[[ $$cld_pr = "gcp" ]] && { terraform init \
		-backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="prefix=terraform/state/$(tf_alias)"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); }; \
	echo "DISPATCHER_VSWITCH_ID is $(DISPATCHER_VSWITCH_ID)"; echo "ori_vs_id is $$ori_vs_id"; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="cloud_pr=$(SLS_CLOUD_PR)" \
		-var="gcp_project=$(GOOGLE_PROJECT)" -var="gcp_region=$(GOOGLE_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="slb_id=$$slb_id" -var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; \
	exit 0

allen_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); \
	[[ $$cld_pr = "ali" ]] && { terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ $$cld_pr = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
		-backend-config="storage_account_name=$(az_storage_account)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="container_name=$(tf_alias)tfstate"; }; \
	init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	netmode="dynamic"; sg_id="none"; vs_id="none"; slb_id="none"; ori_vs_id=$(DISPATCHER_VSWITCH_ID); \
	[[ -n $$ori_vs_id ]] && { netmode="fixed"; sg_id=$(DISPATCHER_SG_ID); vs_id=$(DISPATCHER_VSWITCH_ID); slb_id=$(DISPATCHER_SLB_ID); }; \
	echo "netmode is $$netmode"; echo "sg_id is $$sg_id"; echo "vs_id is $$vs_id"; \
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="slb_id=$$slb_id" -auto-approve; \
	exit 0

mix_install:
	cd $(cur_dir)/dispatcher/$(SLS_CLOUD_PR); cld_pr=$(SLS_CLOUD_PR); \
	[[ $$cld_pr = "ali" ]] && { echo "ali pr"; terraform init -backend-config="bucket=sls-tf-$(oss_alias)" \
		-backend-config="key=$(tf_alias).terraform.tfstate" \
		-backend-config="tablestore_endpoint=https://sls-tf-$(table_alias).$(table_addr)"; }; \
	[[ $$cld_pr = "azure" ]] && { terraform init -backend-config="resource_group_name=$(AZ_RG_NAME)" \
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
	terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" -var="network_mode=$$netmode" \
		-var="slb_id=$$slb_id" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" \
		-var="gcp_project=$(GCP_PROJECT)" -var="gcp_region=$(GCP_REGION)" \
		-var="subnet_ids=$(AZ_SUBNET_IDS)" -var="resource_group_name=$(AZ_RG_NAME)" \
		-var="workspace_id=$(AZ_LOG_ANA_WORKSPACE_ID)" -var="workspace_key=$(AZ_LOG_ANA_WORKSPACE_KEY)" \
		-var="team=$(ralias)" -var="charge_labels=$(WF_SLS_CHARGE_LABELS)" -var="ctx_log_level=$(CTX_LOG_LEVEL)" \
		-var="lazy_regs=$(WF_SLS_REGS)" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var="image_ver=$(image_v)" -var-file="ubuntu_dispatcher.tfvars" \
		-var="network_mode=$$netmode" -var="security_group_id=$$sg_id" -var="vswitch_id=$$vs_id" \
		-var="allen_regs=allen" -var="cloud_pr=$(SLS_CLOUD_PR)" \
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
	echo "duser is $(duser)";
	sudo docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		-f ./image/Dockerfile.runner --build-arg SHR_C_T=runner --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) \
		--build-arg GO_ARCH=amd64 --build-arg SLS_ENC_KEY="$(SLS_ENC_KEY)" \
		--build-arg CTX_USERNAME=$(CTX_USERNAME) --build-arg CTX_PWD=$(CTX_PWD) \
		--build-arg GIT_ACCESS_TOKEN_USR=$(GIT_ACCESS_TOKEN_USR) --build-arg GIT_ACCESS_TOKEN_PWD=$(GIT_ACCESS_TOKEN_PWD) \
		--build-arg RUNNER_PLATFORM=linux-x64 --build-arg DOCKER_USER=$(duser) .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:latest
	echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://gcr.io
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v) \
		gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-runner-eci:$(image_v)
	docker push gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-runner-eci:$(image_v)
	docker tag gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-runner-eci:$(image_v) \
		gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-runner-eci:latest
	docker push gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-runner-eci:latest

dipatcher_image:
	sudo docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=amd64 --build-arg REPO_ORG_NAME=$(REPO_ORG_NAME) \
		--build-arg REPO_ORG_URL=$(REPO_ORG_URL) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg GO_ARCH=amd64 --build-arg GITENT_TK=$(SLS_GITENT_TK) --build-arg GITHUB_TK=$(SLS_GITHUB_TK) \
		--build-arg ALLEN_DB_HOST=$(ALLEN_DB_HOST) --build-arg ALLEN_DB_PORT=$(ALLEN_DB_PORT) --build-arg ALLEN_DB_USR=$(ALLEN_DB_USR) \
		--build-arg ALLEN_DB_PWD=$(ALLEN_DB_PWD) --build-arg ALLEN_DB_DBNAME=$(ALLEN_DB_DBNAME) --build-arg ALLEN_DB_TABLE=$(ALLEN_DB_TABLE) \
		--build-arg SLS_ENC_KEY="$(SLS_ENC_KEY)" \
		--build-arg GIT_ACCESS_TOKEN_USR=$(GIT_ACCESS_TOKEN_USR) --build-arg GIT_ACCESS_TOKEN_PWD=$(GIT_ACCESS_TOKEN_PWD) \
		--build-arg CTX_USERNAME=$(CTX_USERNAME) --build-arg CTX_PWD=$(CTX_PWD) \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:latest
	echo "$$GOOGLE_CREDENTIALS" | docker login -u _json_key --password-stdin https://gcr.io
	docker tag artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v) \
		gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-dispatcher:$(image_v)
	docker push gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-dispatcher:$(image_v)
	docker tag gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-dispatcher:$(image_v) \
		gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-dispatcher:latest
	docker push gcr.io/$(GOOGLE_PROJECT)/serverless-hosted-dispatcher:latest

agent_install:
	cd $(cur_dir)/agent/$(SLS_CLOUD_PR); export ALICLOUD_REGION=$(AGENT_REGION); \
	terraform init; init_code=$$?; echo "init_code is : $$init_code"; \
	[[ $$init_code -ne 0 ]] && terraform init; \
	terraform apply -var-file="ubuntu_agent.tfvars" -auto-approve; apply_code=$$?; echo "apply_code is: $$apply_code"; \
	[[ $$apply_code -ne 0 ]] && terraform apply -var-file="ubuntu_agent.tfvars" -auto-approve; \
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
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=arm64 --build-arg REPO_ORG_NAME=$(REPO_ORG_NAME) \
		--build-arg REPO_ORG_URL=$(REPO_ORG_URL) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) \
		--build-arg GO_ARCH=amd64 --build-arg LOCAL_MODE=True .  
		
local_runner_image:
	sudo docker build \
		-t localhost/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(local_image_v) \
		-f ./image/Dockerfile.runner --build-arg SHR_C_T=runner --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
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
	echo $(image_v)
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
		-f ./image/Dockerfile.runner.remoted --build-arg SHR_C_T=runner --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) \
		--build-arg GO_ARCH=amd64 --build-arg SLS_ENC_KEY="$(SLS_ENC_KEY)" \
		--build-arg RUNNER_PLATFORM=linux-x64 .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci:$(image_v)-rd
dipatcher_image_remoted:
	sudo docker buildx build --platform linux/amd64 \
		-t artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)-rd \
		-f ./image/Dockerfile.dispatcher --build-arg SHR_C_T=dispatcher --build-arg RUNNER_TOKEN=$(TF_VAR_RUNNER_TOKEN) \
		--build-arg MNS_URL=$(TF_VAR_MNS_URL_PUBLIC) --build-arg ACCESS_KEY=$(ALICLOUD_ACCESS_KEY) --build-arg SECRET_KEY=$(ALICLOUD_SECRET_KEY) \
		--build-arg REGION=$(ALICLOUD_REGION) --build-arg TF_VER=1.6.5 --build-arg TF_PLATFORM=amd64 --build-arg REPO_ORG_NAME=$(REPO_ORG_NAME) \
		--build-arg REPO_ORG_URL=$(REPO_ORG_URL) --build-arg IMAGE_RETRIEVE_USERNAME=$(TF_VAR_IMAGE_RETRIEVE_USERNAME)  \
		--build-arg GO_ARCH=amd64 --build-arg GITENT_TK=$(SLS_GITENT_TK) --build-arg GITHUB_TK=$(SLS_GITHUB_TK) \
		--build-arg ALLEN_DB_HOST=$(ALLEN_DB_HOST) --build-arg ALLEN_DB_PORT=$(ALLEN_DB_PORT) --build-arg ALLEN_DB_USR=$(ALLEN_DB_USR) \
		--build-arg ALLEN_DB_PWD=$(ALLEN_DB_PWD) --build-arg ALLEN_DB_DBNAME=$(ALLEN_DB_DBNAME) --build-arg ALLEN_DB_TABLE=$(ALLEN_DB_TABLE) \
		--build-arg SLS_ENC_KEY="$(SLS_ENC_KEY)" \
		--build-arg IMAGE_RETRIEVE_PWD=$(TF_VAR_IMAGE_RETRIEVE_PWD) --build-arg IMAGE_RETRIEVE_SERVER=$(TF_VAR_IMAGE_RETRIEVE_SERVER) .
	docker push artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:$(image_v)-rd