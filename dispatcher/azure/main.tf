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
}