variable eci_group {
    description = "all eci vars definition"
    type = object ({
        name = string
        security_group_id = string
        vswitch_id = string
        image_retrieve_psw = string 
        image_retrieve_server = string
        image_retrieve_uname = string
        security_group_name = optional(string, "serverless-hosted-runner") 
        cpu = optional(string, "1.0") 
        memory = optional(string, "2.0")  
        image_cache = optional(bool, true)  
        restart_policy = optional(string, "OnFailure") 
        tags = optional(map(any), {
            product = "serverless-hosted-runner",
            team = "ccoecn",
            maintainer = "hayue2"
            organization = "ingka-group-digital"
            repository = "serverless-hosted-runner"
            charge_labels = "dind supported"
        })
        add_host_ip = optional(string, "127.0.0.1")
        add_host_fqdn = optional(string, "localhost")
        dns_name_servers = optional(list(string), ["10.82.31.69","10.82.31.116"])
        dns_searches = optional(list(string), ["docker.com","googleapis.com","google.com"])
    })
}

variable eci_container { 
    description = "all eci container vars"
    type = object({
      name = string
      image = string 
      image_ver = string
      cloud_pr = optional(string, "ali")
      ctx_log_level = optional(string, "13")
      container_type = optional(string, "")
      need_privileged = optional(bool, false)
      runner_id = optional(string, "")
      runner_token = optional(string, "")
      runner_repurl = optional(string, "")
      runner_repname = optional(string, "")
      runner_orgname = optional(string, "")
      runner_orgowner = optional(string, "")
      runner_action = optional(string, "")
      runner_lazy_regs = optional(string,"")
      runner_allen_regs = optional(string,"")
      runner_labels = optional(string, "") 
      runner_group = optional(string, "default") 
      image_pull_policy = optional(string, "IfNotPresent")
      working_dir = optional(string, "/tmp")
      startup_cmd = optional(string, "")
      ports_port = optional(string, "80")
      ports_protocol = optional(string, "TCP")
      environment_key = optional(string, "runner")
      environment_val = optional(string, "eci")
      env_docker_host_key = optional(string, "DOCKER_HOST")
      env_docker_host_val = optional(string, "")
      env_docker_tls_verify_key = optional(string, "DOCKER_TLS_VERIFY")
      env_docker_tls_verify_val = optional(string, "")
      volume_mount_name = optional(string, "eci-dockerd-shared-volume-work")
      volume_mount_path = optional(string, "/go/bin/_work")
      volume_mount_name_var_run = optional(string, "eci-dockerd-shared-volume-var-run")
      volume_mount_path_var_run = optional(string, "/var/run")
      volume_mount_name_working_dir = optional(string, "working-dir-volume")
      oss_volume_name = optional(string, "sls-runner-eci-oss-volume") 
      oss_mount_path = optional(string, "/go/bin/_work") 
      oss_bucket = optional(string, "serverless-tfstate") 
      oss_url = optional(string, "oss-cn-shanghai.aliyuncs.com")
      oss_path = optional(string, "/oss_mount")
      oss_ram_role = optional(string, "sls-mount-oss")
    }) 
}
   
