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
    image = var.aci_runner.container_image
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.aci_runner.startup_cmd
    ports_port = var.aci_runner.ports_port 
  }
}