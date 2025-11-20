module "conf" {
    source = "../module/azure/conf"
}

provider "azurerm" {
  features {}
}

module "aci_self_hosted_runner" {
  depends_on = [ module.conf ]
  source = "../module/azure/aci"
  aci_group = { 
    name = var.group_name
    location = var.aci_location
    sku = var.aci_sku
    ip_address_type = var.aci_network_type
    subnet_ids = var.subnet_ids
    resource_group_name = var.resource_group_name
    image_retrieve_server = var.IMAGE_RETRIEVE_SERVER
    image_retrieve_uname = var.IMAGE_RETRIEVE_USERNAME
    image_retrieve_psw = var.IMAGE_RETRIEVE_PWD
    cpu  = var.runner_cpu
    memory = var.runner_memory
    add_host_fqdn = var.add_host_fqdn
    add_host_ip = var.add_host_ip
    tags = {
      product = "serverless-hosted-runner",
      team = "ccoecn",
      maintainer = "hayue2",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
    }
    dns_name_servers = var.dns_name_servers
    dns_searches = var.dns_searches
    restart_policy = var.aci_runner.restart_policy
    workspace_id = var.workspace_id
    workspace_key = var.workspace_key
  }
  aci_container = { 
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
    name  = var.aci_runner.container_name
    # # migrate from artifactory to acr due to timeout issue on azure cn cloud
    # image = var.aci_runner.container_image
    image = join("/", ["${var.IMAGE_RETRIEVE_SERVER}", "ccoecn-docker-virtual/serverless-hosted-runner/serverless-hosted-runner-eci"])
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.aci_runner.startup_cmd
    ports_port = var.aci_runner.ports_port 
    cloud_pr = var.cloud_pr
    dis_ip = var.dis_ip
    repo_reg_tk = var.repo_reg_tk
  }
  aci_container_env_keys = {}
  aci_container_env_vals = {
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