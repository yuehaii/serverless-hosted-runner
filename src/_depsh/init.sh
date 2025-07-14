#!/usr/bin/env bash

[[ -v _INIT_RUNNER ]] && return  
_INIT_RUNNER="$(realpath "${BASH_SOURCE[0]}")"; declare -rg _INIT_RUNNER 

function runner::init::wait_process {
    local max_time_wait=30
    local process_name="$1"
    local waited_sec=0
    while ! pgrep "$process_name" >/dev/null && ((waited_sec < max_time_wait)); do
        echo "process $process_name is not running yet. Retrying in 1 seconds"
        echo "waited $waited_sec seconds of $max_time_wait seconds"
        sleep 1
        ((waited_sec=waited_sec+1))
        if ((waited_sec >= max_time_wait)); then
            return 1
        fi
    done
    return 0
}

function runner::init::dns_server {
    # dc9 dns 
    sudo echo "$(echo 'nameserver 10.82.31.69'; echo 'nameserver 10.82.31.116'; cat /etc/resolv.conf)" > /etc/resolv.conf 
}

function runner::docker:setup { 
    sudo apt-get update
    sudo apt-get install ca-certificates curl
    sudo install -m 0755 -d /etc/apt/keyrings
    sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    sudo chmod a+r /etc/apt/keyrings/docker.asc

    # Add the repository to Apt sources:
    echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
    $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}") stable" | \
    sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update

    sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    sudo docker run hello-world
}

function runner::init::docker_daemon {
    if [ ! -f /etc/docker/daemon.json ]; then
        echo "net.ipv4.ip_forward=1" >> /etc/sysctl.conf
        mkdir -p /etc/docker
        # TODO: ECI container can't be start with privileged, but mount overlay2 need privileged
        #       The overlay dose not need the privileged, but this driver already deprecated by docker
        # echo "{\"storage-driver\": \"overlay2\"}" > /etc/docker/daemon.json
        echo "{\"storage-driver\": \"vfs\"}" > /etc/docker/daemon.json
    fi
    sudo /usr/bin/dockerd > ./dockerd.log 2>&1 &
    processes=(dockerd)
    for process in "${processes[@]}"; do
        if ! runner::init::wait_process "$process"; then
            echo "$process is not running after max time"
            exit 1
        else
            echo "$process is running"
        fi
    done    
}

# runner::docker:setup
runner::init::docker_daemon
runner::init::dns_server