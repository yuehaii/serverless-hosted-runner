variable gcp_group {
    description = "all gcp cloudrun job vars definition"
    type = object ({
        name = optional(string, "serverless-hosted-runner")
        location = optional(string, "us-central1")
        deletion_protection = optional(bool, false)
        ingress = optional(string, "INGRESS_TRAFFIC_ALL")
        min_instance_count = optional(number, 1)
        max_instance_count = optional(number, 5)
        traffic_type = optional(string, "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST")
        traffic_percent = optional(number, 100)
        tags = optional(map(any), {
            product = "serverless-hosted-runner",
            team = "ccoecn",
            maintainer = "hayue2"
            organization = "ingka-group-digital"
            repository = "serverless-hosted-runner"
            charge_labels = "dind supported"
        })
        dns_name_servers = optional(list(string), ["10.82.31.69","10.82.31.116"])
        dns_searches = optional(list(string), ["docker.com","googleapis.com","google.com"])
    })
}

variable gcp_container { 
    description = "all gcp cloudrun job container vars"
    type = object({
      name = optional(string, "serverless-hosted-runner-container")
      image = string 
      image_ver = string
      cpu = optional(string, "1.0") 
      memory = optional(string, "2")  
      ports_name = optional(string, "http1")
      ports_port = optional(string, "80")
      ports_protocol = optional(string, "TCP")
      need_privileged = optional(bool, false) 
      environment_variables_name = optional(string, "runner")
      environment_variables_value = optional(string, "gcp")
      dis_ip = optional(string, "")
      cloud_pr = optional(string, "gcp")
      tf_ctl = optional(string, "go")
      ctx_log_level = optional(string, "13")
      container_type = optional(string, "none")
      runner_id = optional(string, "none")
      runner_token = optional(string, "none")
      runner_repurl = optional(string, "none")
      runner_repname = optional(string, "none")
      runner_orgname = optional(string, "none")
      runner_orgowner = optional(string, "none")
      runner_action = optional(string, "none")
      runner_lazy_regs = optional(string,"none")
      runner_allen_regs = optional(string,"none")
      runner_labels = optional(string, "none") 
      runner_group = optional(string, "default") 
      image_pull_policy = optional(string, "IfNotPresent")
      working_dir = optional(string, "/tmp")
      startup_cmd = optional(string, "none")
      env_docker_host_key = optional(string, "DOCKER_HOST")
      env_docker_host_val = optional(string, "")
      env_docker_tls_verify_key = optional(string, "DOCKER_TLS_VERIFY")
      env_docker_tls_verify_val = optional(string, "none")
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
   
