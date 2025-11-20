variable aci_group {
    description = "all aci vars definition"
    type = object ({
        name = optional(string, "serverless-hosted-runner")
        location = optional(string, "chinanorth3")
        # location = optional(string, "chinaeast2")
        resource_group_name = string
        subnet_ids = optional(string, "")
        # security block available on this sku. but such sku not available in cn region
        # sku = optional(string, "Confidential")
        sku = optional(string, "Standard")
        # ip_address_type = optional(string, "Private")
        ip_address_type = optional(string, "Public")
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
        dns_searches = optional(list(string), ["docker.com","googleapis.com","google.com","registry.terraform.io"])
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
      cpu = optional(string, "1.0") 
      memory = optional(string, "2.0")  
      ports_port = optional(string, "80")
      ports_protocol = optional(string, "TCP")
      need_privileged = optional(bool, false) 
      environment_variables = optional(map(any), {
        runner = "aci", 
      })
      tf_ctl = optional(string, "go") 
      repo_reg_tk = optional(string, "")
      dis_ip = optional(string, "")
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
   
variable aci_container_env_keys {
    description = "aci container environment keys"
    type =  object({
      image_retrieve_server = optional(string, "TF_VAR_IMAGE_RETRIEVE_SERVER")
      image_retrieve_uname = optional(string, "TF_VAR_IMAGE_RETRIEVE_USERNAME")
      image_retrieve_psw = optional(string, "TF_VAR_IMAGE_RETRIEVE_PWD")
      ctx_username = optional(string, "contextusername")
      ctx_pwd = optional(string, "contextpassword")
      var_ctx_username = optional(string, "TF_VAR_CTX_USERNAME")
      var_ctx_pwd = optional(string, "TF_VAR_CTX_PWD")
      kafka_endpoint = optional(string, "KAFKA_INS_ENDPOINT")
      kafka_topic = optional(string, "KAFKA_INS_TOPIC")
      kafka_consumer = optional(string, "KAFKA_INS_CONSUMER")
      kafka_username = optional(string, "KAFKA_INS_USERNAME")
      kafka_pwd = optional(string, "KAFKA_INS_PWD")
      kafka_ca = optional(string, "KAFKA_INS_CA_CERT")
      allan_db_host = optional(string, "ALLEN_DB_HOST")
      allan_db_port = optional(string, "ALLEN_DB_PORT")
      allan_db_usr = optional(string, "ALLEN_DB_USR")
      allan_db_pwd = optional(string, "ALLEN_DB_PWD")
      allan_db_dbname = optional(string, "ALLEN_DB_DBNAME")
      allan_db_table = optional(string, "ALLEN_DB_TABLE")
      git_ent_tk = optional(string, "SLS_GITENT_TK")
      git_hub_tk = optional(string, "SLS_GITHUB_TK")
      enc_key = optional(string, "SLS_ENC_KEY")
      var_enc_key = optional(string, "TF_VAR_SLS_ENC_KEY")
      azure_acr_server = optional(string, "AZURE_ACR_SERVER")
      azure_acr_username = optional(string, "AZURE_ACR_USRNAME")
      azure_acr_pwd = optional(string, "AZURE_ACR_PWD")
      var_azure_acr_server = optional(string, "TF_VAR_AZURE_ACR_SERVER")
      var_azure_acr_username = optional(string, "TF_VAR_AZURE_ACR_USRNAME")
      var_azure_acr_pwd = optional(string, "TF_VAR_AZURE_ACR_PWD")
    })
}

variable aci_container_env_vals {
    description = "aci container environment values"
    type =  object({
      ctx_username_val = string
      ctx_pwd_val = string
      kafka_endpoint_val = string
      kafka_topic_val = string
      kafka_consumer_val = string
      kafka_username_val = string
      kafka_pwd_val = string
      kafka_ca_val = string
      allan_db_host_val = string
      allan_db_port_val = string
      allan_db_usr_val = string
      allan_db_pwd_val = string
      allan_db_dbname_val = string
      allan_db_table_val = string
      git_ent_tk_val = string
      git_hub_tk_val = string
      enc_key_val = string
      azure_acr_server_val = string
      azure_acr_username_val = string
      azure_acr_pwd_val = string
    })
}
