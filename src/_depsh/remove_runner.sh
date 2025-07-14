#!/usr/bin/env bash

[[ -v _REMOVER ]] && return  
_REMOVER="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _REMOVER 

_SRC_PATH="$(dirname $_REMOVER)" 
source "$_SRC_PATH/common.sh"
source "$_SRC_PATH/terraform.sh"

# pri
function remover::runner::_check {
     .; cd $1
    tf::state::refresh; tfstate=$(cat ./terraform.tfstate)
    tfstate_type=$(echo $tfstate | jq .resources[0].type)
    echo "remover::runner::_check, tfstate:"; common::hide::tfstate "$tfstate"; echo "tfstate_type: $tfstate_type"
    if [[ $tfstate_type == *"alicloud_eci_container_group"* || ! -f ./flock_cmd.log ]]; then
        echo "runner already created, or destroyed. skip it."
    else 
        flock_cmd=$(cat ./flock_cmd.log)
        echo "runner creation canceled/failed, create again. flock_cmd is $flock_cmd"
        [[ ! -z $flock_cmd ]] && { common::flock "$flock_cmd"; echo "re-create state $?"; }
    fi     
    popd 
}


function remover::remove::_clean { 
#    echo "waiting tfstate refresh... $(sleep 5)"; \
#       tf::state::refresh; \
#       tfstate=$(common::flock "terraform state list"); \
#       echo "remover::remove::_clean tfstate is --- $tfstate --- "; 
   echo "remover::remove::_clean: $1"; echo "pwd1 is $(pwd)"; tf::state::check "$1"; echo "pwd2 is $(pwd)"
   echo "remover::remove::_clean, rm $1"; echo "pwd is $(pwd)"; rm -rf ./$1
}

# pub
function remover::init {
    common::auth_init $3 $4 $5
    export runner_action=$1 container_name=$2 cloud_pr=$9
    export ARM_CLIENT_ID=${10} ARM_CLIENT_SECRET=${11} ARM_SUBSCRIPTION_ID=${12} \
        ARM_TENANT_ID=${13} ARM_ENVIRONMENT=${14} ARM_RESOURCE_PROVIDER_REGISTRATIONS=${15} \
        AZ_SUBNET_IDS=${16} AZ_LOG_ANA_WORKSPACE_ID=${17} AZ_LOG_ANA_WORKSPACE_KEY=${18}
    [[ -z $container_name ]] && \
        { echo "canceled wf not run on any runner, use $7-$8 instead"; export container_name=$7-$8; }
    echo "action is $runner_action, container name is $container_name"
}

function remover::remove::start {
    [[ -z $container_name || ! -d ./$container_name ]] && \
        { echo "runner dir $container_name dose not exists, skip it."; exit 0; }
    [[ $runner_action == "pool_completed" ]] && sleep 300

    pushd . && cd ./$container_name; rm ./flock_cmd.log
    tf::remove; popd
    remover::remove::_clean "$container_name"
}

function remover::runner {
    repo_runner=$7-$8; org_runner=$6-$8
    [[ $container_name == $repo_runner || $container_name == $org_runner ]] && \
        { echo "ignore self runner check"; exit 0; }
    if [[ -d ./$repo_runner ]]; then 
        echo "remover::runner repo runner: $repo_runner"
        remover::runner::_check $repo_runner; 
    elif [[ -d ./$org_runner ]]; then 
        echo "remover::runner org runner: $org_runner"
        remover::runner::_check $org_runner; 
    else 
        echo "remover::runner not created"
    fi
}

function remover::remove {
    common::print "$@"
    remover::init "$@" 
    remover::remove::start
}

remover::remove "$@"
remover::runner "$@"
