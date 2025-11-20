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
      repo_reg_tk = optional(string, "")
      dis_ip = optional(string, "")
      cloud_pr = optional(string, "ali")
      tf_ctl = optional(string, "go")
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
      docker_volume_name = optional(string, "empty-docker-vm")
      docker_volume_type = optional(string, "EmptyDirVolume")
      docker_mount_path = optional(string, "/var/lib/docker")
    }) 
}

variable eci_mount { 
    description = "all eci mount vars"
    type =  map(object({
      oss_volume_name = optional(string, "sls-runner-eci-oss-volume") 
      oss_mount_path = optional(string, "/go/bin/_work")
      oss_bucket = optional(string, "sls-tf-test") 
      oss_url = optional(string, "oss-cn-shanghai.aliyuncs.com")
      oss_path = optional(string, "/sls_mount")
      oss_ram_role = optional(string, "sls-mount-oss")
      oss_type = optional(string, "FlexVolume")
      oss_driver = optional(string, "alicloud/oss")
    }))
}

variable eci_init_container { 
    description = "all eci init container vars"
    type =  map(object({
      init_container_name = optional(string, "sls-init-container")
      init_container_image = optional(string, "artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher:test-tf-ctl9")
      init_container_pullpolicy = optional(string, "IfNotPresent")
      init_container_cmds = optional(list(string), ["sysctl"])
      init_container_args = optional(list(string), ["-w","vm.overcommit_ratio=10000","-w","vm.overcommit_memory=2"])
    }))
}

variable eci_container_env_keys {
    description = "eci container environment keys"
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

variable eci_container_env_vals {
    description = "eci container environment values"
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