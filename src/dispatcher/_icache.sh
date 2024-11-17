#!/usr/bin/env bash

[[ -v _IMAGE_CACHE ]] && return  
_IMAGE_CACHE="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _IMAGE_CACHE 

function creator::icache::tf_apply {
   echo "creator::icache::tf_apply $(common::print $@)"
   terraform apply -var="vswitch_id=$1" -var="security_group_id=$2" -var="image_ver=$3" \
      -var="username=$4" -var="password=$5" -auto-approve -lock=true &> ./tf_icache_apply.log;
   apply_code=$?; cat ./tf_icache_apply.log; echo "apply_icache_code is : $apply_code"
   [[ $apply_code -ne 0 ]] && { terraform apply -var="vswitch_id=$1" -var="security_group_id=$2" -var="image_ver=$3" \
      -var="username=$4" -var="password=$5" -auto-approve -lock=true &> ./tf_icache_apply.log; }
}
function creator::icache {
   echo "creator::icache $(common::print $@)"
   pushd .
   if [[ "$1" == "create_pool" ]] && [[ ! -d ./${10}/$8 ]]; then
      mkdir -p ./${10}/$8; cp -r ./cache/* ./${10}/$8; cd ./${10}/$8
      creator::tf_init
      creator::icache::tf_apply ${12} ${11} $7 $TF_VAR_IMAGE_RETRIEVE_USERNAME $TF_VAR_IMAGE_RETRIEVE_PWD
   elif [[ "$1" == "queued" ]] && [[ ! -d ./${11}/$9 ]]; then
      mkdir -p ./${11}/$9; cp -r ./cache/* ./${11}/$9; cd ./${11}/$9
      creator::tf_init
      creator::icache::tf_apply ${13} ${12} $8 $TF_VAR_IMAGE_RETRIEVE_USERNAME $TF_VAR_IMAGE_RETRIEVE_PWD
   fi
   popd
}
