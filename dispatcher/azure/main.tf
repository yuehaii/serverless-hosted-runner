module "conf" {
    source = "../../module/azure/conf"
}  

terraform {
  backend "azurerm" {}
}

provider "azurerm" {
  ## https://registry.terraform.io/providers/hashicorp/azurerm/3.116.0/docs
  # skip_provider_registration = true

  ## laest: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs#resource-provider-registrations
  # resource_provider_registrations = "none"
  features {}
}

module "aci_dispatcher_module" {
  depends_on = [ module.conf ]
  source = "../../module/azure/aci"
  aci_group = {
    name = var.aci_dispatcher.group_name
    resource_group_name = var.resource_group_name
    subnet_ids = var.subnet_ids
    image_retrieve_server = var.IMAGE_RETRIEVE_SERVER
    image_retrieve_uname = var.IMAGE_RETRIEVE_USERNAME
    image_retrieve_psw = var.IMAGE_RETRIEVE_PWD
    image_cache = true
    tags = {
      product = "serverless-hosted-runner",
      team = var.team,
      maintainer = "hayue2",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
    }
    workspace_id = var.workspace_id
    workspace_key = var.workspace_key
  }
  aci_container = {
    name = var.aci_dispatcher.container_name
    image = var.aci_dispatcher.container_image
    cpu  = var.dispacher_cpu
    memory = var.dispacher_memory
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.aci_dispatcher.startup_cmd
    ports_port = var.aci_dispatcher.ports_port 
    container_type = var.container_type
    runner_id = var.runner_id
    runner_token = var.runner_token
    runner_repurl = var.runner_repurl
    runner_repname = var.runner_repname
    runner_action = var.runner_action
    runner_orgname = var.runner_orgname
    runner_orgowner = var.runner_orgowner
    runner_lazy_regs = var.lazy_regs
    runner_allen_regs = var.allen_regs
    cloud_pr = var.cloud_pr
    tf_ctl = var.tf_ctl
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