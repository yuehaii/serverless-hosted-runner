module "conf" {
    source = "../module/ali/conf"
}

module "eci_net" {
    source = "../module/ali/net"
    net_var = {
        sg_name = var.eci_runner.security_group_name
    }
    count = var.network_mode == "dynamic" ? 1 : 0
}
 
module "eci_nat" {
    depends_on = [ module.conf, module.eci_net ]
    source = "../module/ali/nat"
    net_var = {
        vswitch_id = module.eci_net[0].net_vswitch_id
        vpc_id = module.eci_net[0].net_vpc_id 
    }
    count = var.network_mode == "dynamic" ? 1 : 0
} 

module "eci_self_hosted_runner" {
  depends_on = [ module.eci_net, module.eci_nat ]
  source = "../module/ali/eci"
  eci_group = { 
    name = var.group_name
    image_retrieve_server = var.IMAGE_RETRIEVE_SERVER
    image_retrieve_uname = var.IMAGE_RETRIEVE_USERNAME
    image_retrieve_psw = var.IMAGE_RETRIEVE_PWD
    security_group_id = var.security_group_id == "" ? module.eci_net[0].net_sg_id : var.security_group_id
    vswitch_id = var.vswitch_id == "" ? module.eci_net[0].net_vswitch_id : var.vswitch_id
    cpu  = var.runner_cpu
    memory = var.runner_memory
    add_host_fqdn = var.add_host_fqdn
    add_host_ip = var.add_host_ip
    tags = {
      product = "serverless-hosted-runner",
      team = "ccoecn",
      maintainer = "hayue2",
      "k8s.aliyun.com/eci-docker-build-enable" = "true",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
    }
    dns_name_servers = var.dns_name_servers
    dns_searches = var.dns_searches
    restart_policy = var.eci_runner.restart_policy
  }
  eci_container = { 
    container_type = var.container_type
    need_privileged = true
    runner_id = var.runner_id
    runner_token = var.runner_token
    runner_repurl = var.runner_repurl
    runner_repname = var.runner_repname
    runner_action = var.runner_action
    runner_orgname = var.runner_orgname
    runner_orgowner = var.runner_orgowner
    runner_labels = var.runner_labels
    runner_group = var.runner_group
    name  = var.eci_runner.container_name
    image = var.eci_runner.container_image
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    working_dir = var.eci_runner.working_dir
    startup_cmd = var.eci_runner.startup_cmd
    ports_port = var.eci_runner.ports_port 
    cloud_pr = var.cloud_pr
    dis_ip = var.dis_ip
    repo_reg_tk = var.repo_reg_tk
  }
  eci_mount = var.oss_mount == ""? {} : {
    "oss-sls" =  {
      oss_mount_path = "/go/bin/_work"
      oss_volume_name = var.runner_repname
      oss_mount_path = "/go/bin/_work"
      oss_bucket = var.oss_mount
      oss_url = "oss-cn-shanghai.aliyuncs.com"
      oss_path =  join("/", ["/sls_mount", var.runner_repname, var.runner_id, var.oss_mount])
      oss_ram_role = "sls-mount-oss"
      oss_type = "FlexVolume"
      oss_driver = "alicloud/oss"
    }
  }
  eci_init_container = {}
  eci_container_env_keys = {}
  eci_container_env_vals = {
    ctx_username_val = var.CTX_USERNAME
    ctx_pwd_val = var.CTX_PWD
    kafka_endpoint_val = var.KAFKA_INS_ENDPOINT
    kafka_topic_val = var.KAFKA_INS_TOPIC
    kafka_consumer_val = var.KAFKA_INS_CONSUMER
    kafka_username_val = var.KAFKA_INS_USERNAME
    kafka_pwd_val = var.KAFKA_INS_PWD
    kafka_ca_val = var.KAFKA_INS_CA_CERT
    allan_db_host_val = var.ALLEN_DB_HOST
    allan_db_port_val = var.ALLEN_DB_PORT
    allan_db_usr_val = var.ALLEN_DB_USR
    allan_db_pwd_val = var.ALLEN_DB_PWD
    allan_db_dbname_val = var.ALLEN_DB_DBNAME
    allan_db_table_val = var.ALLEN_DB_TABLE
    git_ent_tk_val = var.SLS_GITENT_TK
    git_hub_tk_val = var.SLS_GITHUB_TK
    enc_key_val = var.SLS_ENC_KEY
    azure_acr_server_val = var.AZURE_ACR_SERVER
    azure_acr_username_val = var.AZURE_ACR_USRNAME
    azure_acr_pwd_val = var.AZURE_ACR_PWD
  }
}