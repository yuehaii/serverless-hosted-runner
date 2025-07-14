#!/usr/bin/env bash

[[ -v _TERRAFORM ]] && return  
_TERRAFORM="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _TERRAFORM

_SRC_PATH="$(dirname $_TERRAFORM)" 
source "$_SRC_PATH/common.sh"

# pri
function tf::_paths {
   echo "##current dir $(pwd)"
   echo "##files in . dir $(ls -al)"
   echo "##files in .. dir $(ls -al ../)"
   echo "##files in ../../ dir $(ls -al ../../)"
   echo "##files in /go/bin/tf_plugin_cache dir $(ls -al /go/bin/tf_plugin_cache)"
   echo "##files in /go/bin/runner dir $(ls -al /go/bin/runner/)"
   echo "##files in /go/bin dir $(ls -al /go/bin/)"
   echo "##files in /go dir $(ls -al /go/)"
}

function tf::init::_cmd {
   # tf::_paths
   flock_cmd="terraform init $@"
   common::flock "$flock_cmd"
}

function tf::apply::_cmd {
   echo "tf::apply::_cmd $@"
   flock_cmd="terraform apply -var="runner_action=$runner_action" \
-var="runner_repurl=$runner_url" -var="runner_token=$runner_pat" \
-var="image_ver=$runner_ver" -var="security_group_id=$runner_sg" \
-var="vswitch_id=$runner_vswith" -var="container_type=$runner_type" \
-var="runner_orgname=$runner_org" -var="network_mode=$network_mode" \
-var="runner_cpu=$runner_cpu" -var="runner_memory=$runner_memory" \
-var="runner_labels=$runner_labels" -var="charge_labels=$runner_charge_labels" \
-var="runner_group=$runner_group" -var="ctx_log_level=$ctx_log_level" \
-var="subnet_ids=$arm_subnet_id" -var="resource_group_name=$arm_resource_group_name" \
-var="workspace_id=$arm_log_ana_workspace_id" -var="workspace_key=$arm_log_ana_workspace_key" \
-var="gcp_project=$GOOGLE_PROJECT" -var="gcp_region=$GOOGLE_REGION" \
-var="gcp_project_sa_email=$GOOGLE_PROJECT_SA_EMAIL" -var="gcp_project_apikey=$GOOGLE_PROJECT_APIKEY" -var="gcp_runner_dind=$GOOGLE_RUNNER_DIND" \
-var="gcp_vpc=$gcp_vpc" -var="gcp_subnet=$gcp_subnet" \
-var="add_host_ip=47.76.42.71" -var="add_host_fqdn=serverless.dockerd.com" "$@" \
-var-file="ubuntu_runner.tfvars" -auto-approve"
   echo $flock_cmd > flock_cmd.log; echo "tf::apply::_cmd flock_cmd: $(common::hide::tfcmd $flock_cmd)"
   common::flock "$flock_cmd"
}

## Ali tf bug
# Env: smoke test with 0.25vcore/0.5M and 4 concurrent tf apply 
# Des: tf apply canceled the request: "The plugin.(*GRPCProvider).ApplyResourceChange request was cancelled."
#      and it triggered the retry. but the first tf apply not been canceled in ali server side. 
#      so finally the first tf apply created the ECI normally and second retry also create a duplicate one.
# Workaround (done): work aroud code as below. 
#      if still met such issue for other tf apply states, will disable retry (disabled).
# Solution: ali tf team need to fix the issue. 
#    
function tf::state::_clean {
   [[ "$1" == *"ApplyResourceChange request was cancelled"* ]] && \
      { echo "tf_apply::state_clean"; export apply_code=0; echo "cleaned apply code is $apply_code"; }
}

# not safe if tf state format change. need it under high workload
function tf::state::check::_ps {
   pushd . ; cd $1; tf::state::refresh; popd; tfstate=$(cat $1/terraform.tfstate)
   # echo "tf::state::check::ps: $tfstate."
   tfstate_type=$(echo $tfstate | jq .resources[0].type)
   if [[ $tfstate_type == *"alicloud_eci_container_group"* || $tfstate_type == *"azurerm_container_group"* || $tfstate_type == *"google_cloud_run"* || $tfstate_type == *"gcp_runner_batch_job_module"* ]]; then
      echo "_ps runner already occupied"; exit 0;
   else 
      echo "_ps runner not created yet, destroyed or canceled."
   fi     
}

# safe. but consume more cpu resource
function tf::state::check::_tf {
   pushd . ; cd $1; tfstate=$(common::flock "terraform state list"); 
   # tf_f=$(cat ./terraform.tfstate); echo "tf_f: $tf_f." 
   echo "tf::state::check::_tf: $tfstate. "
   if [[ $tfstate == *"alicloud_eci_container_group"* || $tfstate == *"azurerm_container_group"* || $tfstate == *"google_cloud_run"* || $tfstate == *"gcp_runner_batch_job_module"* ]]; then
      echo "_tf runner already occupied"; exit 0;
   else 
      echo "_tf runner not created yet, destroyed or canceled."      
      echo "pwd3 is $(pwd)"; popd; echo "pwd4 is $(pwd)"
   fi     
}

function tf::destroy::_occupied {
   echo "remove occupied runner. need rerun job with runner not response."
   terraform destroy -var-file="ubuntu_runner.tfvars" -auto-approve
}

function tf::remove::_lock { 
    if [[ "$cloud_pr" == "azure" ]]; then
      echo "in azure destory"
      common::flock "terraform destroy -var-file=ubuntu_runner.tfvars -var="subnet_ids=$AZ_SUBNET_IDS" -var="workspace_id=$AZ_LOG_ANA_WORKSPACE_ID" -var="workspace_key=$AZ_LOG_ANA_WORKSPACE_KEY" -auto-approve $@"
    else 
      common::flock "terraform destroy -var-file=ubuntu_runner.tfvars -auto-approve $@"
    fi
}

# pub 
function tf::init {
   echo "creator::tf_init $(common::print $@)"
   tf::init::_cmd &> ./tf_init.log ; init_code=$?; 
   cat ./tf_init.log; echo "init_code is : $init_code";
   [[ $init_code -ne 0 ]] && { tf::init::_cmd; cat ./tf_init.log.retry; }
}
  
function tf::apply {
   echo "tf::apply $(common::print $@)"
   if [[ "$runner_action" == "create_pool" ]]; then
      tf::apply::_cmd &> ./tf_apply.log; \
         apply_code=$?; tf_app=$(cat ./tf_apply.log); echo "tf_apply_log is:"; common::hide::tfstate "$tf_app"; echo "apply_code is : $apply_code";
   elif [[ "$runner_action" == "queued" ]]; then 
      tf::apply::_cmd -var="runner_id=$runner_id" \
         -var="runner_repname=$runner_name" -var="runner_orgowner=$runner_owner" &> ./tf_apply.log; \
         apply_code=$?; tf_apply_log=$(cat ./tf_apply.log); echo "tf_apply_log is:"; common::hide::tfstate "$tf_apply_log"; echo "apply_code is : $apply_code";
   fi
   tf::response
}

function tf::cache::init {
   if [[ ! -d /go/bin/tf_plugin_cache ]]; then
      # tf::_paths
      echo "tf::cache::init for $queued_container_name"
      mkdir /go/bin/tf_plugin_cache && export TF_PLUGIN_CACHE_DIR="/go/bin/tf_plugin_cache" 
      pushd . && cd /go/bin/runner/${cloud_pr} &&  terraform init
      cp ./.terraform.lock.hcl /go/bin/tf_plugin_cache/.terraform.lock.hcl
      rm -rf ./.terraform && rm ./.terraform.lock.hcl; popd
   fi
}

function tf::cache::create {
   export TF_PLUGIN_CACHE_DIR="/go/bin/tf_plugin_cache"
   cp /go/bin/tf_plugin_cache/.terraform.lock.hcl $1
}

function tf::state::check {
   # TODO: found in allen db integration. 
   #       'terraform state list' return null (high cpu or ali tf provider issue), 
   #       but tf state file recorded. 
   # workaround : merge tf and ps method
   # TODO: if terrform.state dose not exist. it will cause process exist in GCP
   if [[ -d $1 && -f $1/terraform.tfstate ]]; then
      tf::state::check::_ps $1 
      tf::state::check::_tf $1 
   else
      echo "tf::state::check, terraform.tfstate not created."
   fi
}

function tf::state::refresh {
   terraform refresh -var-file=ubuntu_runner.tfvars
}

function tf::remove {
   # TODO: When CPU/memory reached to 100% for long time during smoke test, the ali tf provider would cancel
   #       the request with error "The plugin.(*GRPCProvider).ApplyResourceChange request was cancelled.".
   #       But the ECI still created in ali cloud and nothing recorded in tf state file. 
   #       So tf can't delete such canceled service.
   # Solution: need ali team to fix such issue.
   # Workaround: dispatcher need to scan such zombie eci and make clean (done. shutdown such eci to save cost.).
   echo "tf::remove exec";
   rm_code=-1; tf::remove::_lock &> ./tf_rm.log; rm_code=$?;
   tf_app=$(cat ./tf_rm.log); common::hide::tfstate "tf_rm.log is ---- $tf_app ----"; echo "rm_code is : $rm_code"
   [[ $rm_code -ne 0 ]] && { rm_retry_code=-1; tf::remove::_lock &> ./tf_rm.log.retry; rm_retry_code=$?; \
      tf_r_app=$(cat ./tf_rm.log.retry); common::hide::tfstate "tf_rm.log.retry is ---- $tf_r_app ----"; echo "rm_retry_code is : $rm_retry_code"; }
   [[ $rm_retry_code -ne 0 && $rm_code -ne 0 ]] && { echo "tf::remove failed"; exit 0; }
   
}

function tf::response {
   echo "tf::response, cloud_pr is $cloud_pr, GOOGLE_RUNNER_DIND is $GOOGLE_RUNNER_DIND"
   if [[ "$cloud_pr" == "gcp" && "$GOOGLE_RUNNER_DIND" == "true" ]]; then
      echo "response code is: $(cat ./http_response_code.log)"
      echo "response body is: $(cat ./http_response_body.log)"
   fi
}