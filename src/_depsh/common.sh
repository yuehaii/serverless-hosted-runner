#!/usr/bin/env bash

[[ -v _COMMON ]] && return  
_COMMON="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _COMMON 

function common::print {
   for para in "$@"; do echo "$(common::hide $para), "; done
}

function common::hide {
   sed 's/.\{'${2:-5}'\}$/*****/' <<< $1
}

function common::hide::tfstate {
   sed -E 's/password.{200}/password=*****/g' <<< $1 | sed -E 's/private.{80}/private=**********/g'
}

function common::hide::tfcmd {
   sed -E 's/runner_token=.{20}/runner_token=*****/g' <<< $1
}

function common::auth_init {
   export ALICLOUD_ACCESS_KEY=$1 ALICLOUD_SECRET_KEY=$2 ALICLOUD_REGION=$3
}

function common::flock {
   flock -n -E 37 ./ -c "$1"
}

function common::log::log {
   local color instant level

   color=${1:?missing required <color> argument}
   shift

   level=${FUNCNAME[1]}
   level=${level#log.}
   level=${level^^}

   if [[ ! -v "LOG_${level}_DISABLED" ]]; then
      instant=$(date '+%F %T.%-3N' 2>/dev/null || :)

      # https://no-color.org/
      if [[ -v NO_COLOR ]]; then
         printf -- '%s  %s --- %s\n' "$instant" "$level" "$*" 1>&2 || :
      else
         printf -- '\033[0;%dm%s  %s --- %s\033[0m\n' "$color" "$instant" "$level" "$*" 1>&2 || :
      fi
   fi
}

function common::log::debug { 
   common::log::log 37 "$@"
}

function common::log::notice { 
   common::log::log 34 "$@"
}

function common::log::warning {
   common::log::log 33 "$@"
}

function common::log::error { 
   common::log::log 31 "$@"
}

function common::log::success { 
   common::log::log 32 "$@"
}