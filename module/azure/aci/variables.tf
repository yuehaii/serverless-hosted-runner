variable aci_group {
    description = "all aci vars definition"
    type = object ({
        name = optional(string, "serverless-hosted-runner")
        location = optional(string, "chinanorth3")
        # location = optional(string, "chinaeast2")
        resource_group_name = string
        subnet_ids = string
        # security block available on this sku. but such sku not available in cn region
        sku = optional(string, "Confidential") 
        # testing on cn cloud
        # sku = optional(string, "Standard")
        ip_address_type = optional(string, "Private")
        # ip_address_type = optional(string, "Public")
        # DNS label/name is not supported when deploying to virtual networks.
        dns_name_label = optional(string, "aci-label")
        os_type = optional(string, "Linux")
        image_retrieve_psw = string 
        image_retrieve_server = string
        image_retrieve_uname = string
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
        workspace_id = optional(string, "none")
        workspace_key = optional(string, "none")
    })
}

variable aci_container { 
    description = "all aci container vars"
    type = object({
      name = optional(string, "serverless-hosted-runner-container")
      image = string 
      image_ver = string
      cpu = optional(string, "1") 
      memory = optional(string, "2")  
      ports_port = optional(string, "80")
      ports_protocol = optional(string, "TCP")
      need_privileged = optional(bool, false) 
      environment_variables = optional(map(any), {
        runner = "aci", 
      })
      cloud_pr = optional(string, "azure")
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
   
