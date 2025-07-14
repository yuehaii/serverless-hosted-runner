#!/usr/bin/env bash

[[ -v _SACRIFY_THY ]] && return  
_SACRIFY_THY="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _SACRIFY_THY 

function runner::sacrify::satisfied {
    declare -i idle_time=$1; declare -i MAX_IDLE_TIME=300
    local sacrify_process=1
    local job_complete="[J|j]ob completed"
    local job_progress="[J|j]ob message"
    local job_state="$(cat _diag/Worker_*.log | grep "$job_complete")"
    echo "runner::sacrify, job_state:$job_state"
    [[ $job_state =~ $job_complete ]] && \
        { echo "sacrify satisfied, job_state:$job_state"; \
            kill $sacrify_process; exit 0; }
    
    job_state="$(cat _diag/Worker_*.log | grep "$job_progress")"
    [[ ! $job_state =~ $job_progress && $idle_time -ge $MAX_IDLE_TIME ]] && \
        { echo "sacrify satisfied, job_state:$job_state, idle_time:$idle_time"; \
            kill $sacrify_process; exit 0; }
    echo "sacrify not satisfied, job_state:$job_state"
}

runner::sacrify::satisfied $1