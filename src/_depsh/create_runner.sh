#!/usr/bin/env bash

[[ -v _CREATOR ]] && return  
_CREATOR="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _CREATOR 

_SRC_PATH="$(dirname $_CREATOR)" 
source "$_SRC_PATH/common.sh"
source "$_SRC_PATH/terraform.sh"
source "$_SRC_PATH/access.sh"

# pri
function creator::paras::_pool {
   export runner_action=$1 runner_pool_name=$2 runner_pool_size=$3 runner_org=$4 \
               runner_url=$5 runner_pat=$6 runner_ver=$7 runner_key=$8 runner_sec=$9 \
               runner_region=${10} runner_sg=${11} runner_vswith=${12} runner_type=${13} \
               runner_cpu=${14} runner_memory=${15} runner_labels=${16} runner_charge_labels=${17} \
               runner_group=${18} ctx_log_level=${19} cloud_pr=${20} \
               ARM_CLIENT_ID=${21} ARM_CLIENT_SECRET=${22} ARM_SUBSCRIPTION_ID=${23} \
               ARM_TENANT_ID=${24} ARM_ENVIRONMENT=${25} ARM_RESOURCE_PROVIDER_REGISTRATIONS=${26} \
               arm_resource_group_name=${27} arm_subnet_id=${28} \
               arm_log_ana_workspace_id=${29} arm_log_ana_workspace_key=${30} \
               GOOGLE_CREDENTIALS=$(echo ${31} | base64 -d) GOOGLE_PROJECT=${32} GOOGLE_REGION=${33} \
               GOOGLE_PROJECT_SA_EMAIL=${34} GOOGLE_CREDENTIALS_PRIVATEKEY=$(echo ${35} | base64 -d) GOOGLE_RUNNER_DIND=${36} \
               gcp_vpc=${37} gcp_subnet=${38} \
               network_mode=fixed; \
         echo "action: $1, pool name: $2, pool size: $3, org name: $4, org url: $5"; \
         echo "pat: $(common::hide $6), ver: $7, key: $(common::hide $8), sec: $(common::hide $9), region: ${10}, sg: ${11}, vswith: ${12}, type: ${13}"; \
         echo "cpu: ${14}, memory: ${15}, labels: ${16}, charge_labels: ${17}"; \
         echo "GOOGLE_PROJECT: ${GOOGLE_PROJECT}, GOOGLE_REGION: ${GOOGLE_REGION}"; \
         echo "GOOGLE_CREDENTIALS: $(common::hide $GOOGLE_CREDENTIALS 50)"
   export GOOGLE_PROJECT_APIKEY=$(access::token https://www.googleapis.com/auth/cloud-platform)
   echo "GOOGLE_PROJECT_APIKEY: $(common::hide $GOOGLE_PROJECT_APIKEY 30)"
}

function creator::paras::_ephemeral {
   export runner_action=$1 runner_id=$2 runner_name=$3 runner_url=$4 \
            runner_owner=$6 runner_pat=$7 runner_ver=$8 runner_key=$9 \
            runner_sec=${10} runner_region=${11} runner_sg=${12} runner_vswith=${13} \
            runner_type=${14} runner_cpu=${15} runner_memory=${16} runner_labels=${17} \
            runner_charge_labels=${18} runner_group=${19} ctx_log_level=${20} cloud_pr=${21} \
            ARM_CLIENT_ID=${22} ARM_CLIENT_SECRET=${23} ARM_SUBSCRIPTION_ID=${24} \
            ARM_TENANT_ID=${25} ARM_ENVIRONMENT=${26} ARM_RESOURCE_PROVIDER_REGISTRATIONS=${27} \
            arm_resource_group_name=${28} arm_subnet_id=${29} \
            arm_log_ana_workspace_id=${30} arm_log_ana_workspace_key=${31} \
            GOOGLE_CREDENTIALS=$(echo ${32} | base64 -d) GOOGLE_PROJECT=${33} GOOGLE_REGION=${34} \
            GOOGLE_PROJECT_SA_EMAIL=${35} GOOGLE_CREDENTIALS_PRIVATEKEY=$(echo ${36} | base64 -d) GOOGLE_RUNNER_DIND=${37} \
            gcp_vpc=${38} gcp_subnet=${39} \
            network_mode=fixed; \
        export runner_org="none"; [[ ${runner_type} == "org" ]] && export runner_org=$5; \
        queued_container_name="$3-$2"; \
            [[ $runner_type == "org" ]] && queued_container_name="$5-$2";
        export GOOGLE_PROJECT_APIKEY=$(access::token https://www.googleapis.com/auth/cloud-platform)
        echo "action: $1, id: $2, repo name: $3, repo url: $4, org name: $5, owner: $6, pat: $(common::hide $7), ver: $8"; \
        echo "key: $(common::hide $9), sec: $(common::hide ${10}), region: ${11}, sg: ${12}, vswith: ${13}, type: ${14}"; \
        echo "cpu: ${15}, memory: ${16}, labels: ${17}, charge_labels: ${18}, group is $3-$2"; \
        echo "GOOGLE_CREDENTIALS: $(common::hide $GOOGLE_CREDENTIALS 50)"; \
        echo "GOOGLE_PROJECT_APIKEY: $(common::hide $GOOGLE_PROJECT_APIKEY 30), --END-- GOOGLE_PROJECT: ${GOOGLE_PROJECT}, GOOGLE_REGION: ${GOOGLE_REGION}";
}

function creator::pool::_create {
   cur_dir=$(pwd)
   declare -i num=$runner_pool_size
   mkdir ${cur_dir}/module/${cloud_pr}/$runner_pool_name; cp -r ${cur_dir}/module/${cloud_pr}/eci/* ${cur_dir}/module/${cloud_pr}/$runner_pool_name/
   cp ${cur_dir}/template/${cloud_pr}/eci.pool.pre.tpl ${cur_dir}/module/${cloud_pr}/$runner_pool_name/main.tf
   mkdir ${cur_dir}/$runner_pool_name; cp -r ${cur_dir}/runner/${cloud_pr}/* ${cur_dir}/$runner_pool_name/; \
      export pool_module_path="../module/${cloud_pr}/$runner_pool_name" 
   envsubst '${pool_module_path}' < ${cur_dir}/template/${cloud_pr}/eci.pool.runner.tpl > ${cur_dir}/$runner_pool_name/main.tf
   for (( size=1; size<=$num; size++ ))
   do  
      export pool_container_name=$runner_org"-runner-"$size container_id=$size
      echo "add # $size pool, name is $pool_container_name"
      envsubst '${pool_container_name}, ${container_id}' < ${cur_dir}/template/${cloud_pr}/eci.pool.container.tpl > ./container_tmp.tf 
      cat ${cur_dir}/module/${cloud_pr}/$runner_pool_name/main.tf ./container_tmp.tf >> ${cur_dir}/module/${cloud_pr}/$runner_pool_name/main.tf; \
         rm ./container_tmp.tf
   done 
   cat ${cur_dir}/module/${cloud_pr}/$runner_pool_name/main.tf ${cur_dir}/template/${cloud_pr}/eci.pool.append.tpl >> ${cur_dir}/module/${cloud_pr}/$runner_pool_name/main.tf
   tf::cache::init
   tf::cache::create ${cur_dir}/$runner_pool_name; cd ${cur_dir}/$runner_pool_name
}

function creator::ephemeral::_create {
   cur_dir=$(pwd)
   [[ -z $queued_container_name || -f ${cur_dir}/$queued_container_name/.terraform.tfstate.lock.info ]] && \
      { echo "container group $queued_container_name is empty, or .terraform.tfstate.lock.info exists. skip it."; exit 0; }
   [[ -d ${cur_dir}/$queued_container_name ]] && \
      { tf::state::check $cur_dir/$queued_container_name; }
   cp -r ${cur_dir}/runner/${cloud_pr} ${cur_dir}/$queued_container_name; cd ${cur_dir}/$queued_container_name 
   tf::cache::init
   tf::cache::create ./
}

# pub
function creator::paras {
   [[ "$1" == "create_pool" ]] && { creator::paras::_pool "$@"; }
   [[ "$1" == "queued" ]] && { creator::paras::_ephemeral "$@"; }
}

function creator::init {
   if [[ "$runner_action" == "create_pool" ]]; then
      creator::pool::_create
   elif [[ "$runner_action" == "queued" ]]; then
      creator::ephemeral::_create
   fi 
}

function creator::create {
   cur_dir=$(pwd) 
   creator::paras "$@" 
   if [[ "$runner_action" == "create_pool" && $runner_pool_size -gt 0 && ! -d ${cur_dir}/$runner_pool_name ]] \
      || [[ "$runner_action" == "queued" ]]; then 
      common::auth_init $runner_key $runner_sec $runner_region
      creator::init
      tf::init
      tf::apply
   fi 
   cd $cur_dir
}

creator::create "$@"