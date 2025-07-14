variable batch_job {
    description = "all gcp batch job vars definition"
    type = object ({
        name = optional(string, "serverless-hosted-runner")
        provider = optional(string, "google-batch-v1")
        project_id = string
        sa_email = string
        api_key = string # api key is not allow for batch auth, this value will be set as jwt token signed by sa.
        vpc_name = string
        subnet_name = string
        location = optional(string, "us-central1")
        rest_prefix = optional(string, "https://batch.googleapis.com/v1")
        tags = optional(map(any), {
            product = "serverless-hosted-runner",
            team = "ccoecn",
            maintainer = "hayue2"
            organization = "ingka-group-digital"
            repository = "serverless-hosted-runner"
            charge_labels = "dind supported"
        })
    })
}

variable batch_job_container { 
    description = "all gcp cloudrun job container vars"
    type = object({
        name = optional(string, "serverless-hosted-runner-container")
        image = string 
        image_retrieve_psw = string 
        image_retrieve_uname = string
        image_ver = string
        cpu = optional(string, "2") //2vCpu
        memory = optional(string, "2") //2GB 
        extra_disk = optional(string, "0") //0Mb
        max_run_duration = optional(string, "18000s")
        ports_name = optional(string, "http1")
        ports_port = optional(string, "80")
        ports_protocol = optional(string, "TCP")
        need_privileged = optional(bool, false) 
        environment_variables_name = optional(string, "runner")
        environment_variables_value = optional(string, "gcp")
        dis_ip = optional(string, "")
        tf_ctl = optional(string, "go")
        cloud_pr = optional(string, "gcp")
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
        // tried --cap-add=CAP_NET_ADMIN --cap-add=CAP_SYS_ADMIN to reduce the permission
        optional_paras = optional(string, "--privileged")
    }) 
}
   
