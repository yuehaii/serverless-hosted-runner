### need pre-create rg, vnet/subnet
# resource "azurerm_resource_group" "serverless_rg_template" {
#   name     = "aci-rg"
#   location = "chinanorth3"
# }

locals {  
    base_cmds = var.aci_container.startup_cmd == "" ? [] : ["${var.aci_container.startup_cmd}"]  
    cmds = var.aci_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.aci_container.image_ver, "-r", var.aci_container.runner_lazy_regs, "-a", var.aci_container.runner_allen_regs, 
        "-m", var.aci_container.ctx_log_level, "-c", var.aci_container.cloud_pr]) : concat(local.base_cmds, 
        ["-t", var.aci_container.container_type, "-i", var.aci_container.runner_id, "-k", var.aci_container.runner_token, 
        "-l", var.aci_container.runner_repurl, "-n", var.aci_container.runner_repname, "-a", var.aci_container.runner_action, 
        "-o", var.aci_container.runner_orgname, "-p", var.aci_container.runner_orgowner, "-v", var.aci_container.image_ver,
        "-b", var.aci_container.runner_labels, "-g", var.aci_container.runner_group, "-m", var.aci_container.ctx_log_level]) 
    liveness_probe_period_seconds        = "10"
    liveness_probe_initial_delay_seconds = "5"
    liveness_probe_success_threshold     = "1"
    liveness_probe_failure_threshold     = "1000"
    liveness_probe_timeout_seconds       = "8"
    liveness_probe_cmds                  = ["pwd"]   

    readiness_probe_period_seconds        = "10"
    readiness_probe_initial_delay_seconds = "5"
    readiness_probe_success_threshold     = "1"
    readiness_probe_failure_threshold     = "1000"
    readiness_probe_timeout_seconds       = "8"
    readiness_probe_cmds                  = ["pwd"]     
}

resource "azurerm_container_group" "serverless_aci_template" {
  name                = var.aci_group.name
  location            = var.aci_group.location
  sku                 = var.aci_group.sku
  resource_group_name = var.aci_group.resource_group_name
  ip_address_type     = var.aci_group.ip_address_type
  # dns_name_label      = var.aci_group.dns_name_label
  os_type             = var.aci_group.os_type
  restart_policy      = var.aci_group.restart_policy
  tags                = var.aci_group.tags
  image_registry_credential {
    username = var.aci_group.image_retrieve_uname
    password = var.aci_group.image_retrieve_psw
    server = var.aci_group.image_retrieve_server
  }
  subnet_ids          = [var.aci_group.subnet_ids]
  dns_config {
    nameservers = var.aci_group.dns_name_servers
    search_domains = var.aci_group.dns_searches
    options = ["edns0"]
  }
  diagnostics {
    log_analytics {
      workspace_id = var.aci_group.workspace_id
      workspace_key = var.aci_group.workspace_key
    }
  }
  container {
    name   = var.aci_container.name
    image = join(":", [var.aci_container.image, var.aci_container.image_ver])
    cpu    = var.aci_container.cpu
    memory = var.aci_container.memory
    commands = local.cmds
    ports {
        port     = var.aci_container.ports_port
        protocol = var.aci_container.ports_protocol
    }
    # TODO: need this feature for dind solution
    # security {
    #     privilege_enabled = var.aci_container.need_privileged
    # }
    environment_variables = var.aci_container.environment_variables
    
    liveness_probe {
        period_seconds        = local.liveness_probe_period_seconds
        initial_delay_seconds = local.liveness_probe_initial_delay_seconds
        success_threshold     = local.liveness_probe_success_threshold
        failure_threshold     = local.liveness_probe_failure_threshold
        timeout_seconds       = local.liveness_probe_timeout_seconds
        exec                  = local.liveness_probe_cmds 
    }
    readiness_probe {
        period_seconds        = local.readiness_probe_period_seconds
        initial_delay_seconds = local.readiness_probe_initial_delay_seconds
        success_threshold     = local.readiness_probe_success_threshold
        failure_threshold     = local.readiness_probe_failure_threshold
        timeout_seconds       = local.readiness_probe_timeout_seconds
        exec                  = local.readiness_probe_cmds 
    }
  }
}

output "aci_id" {
  value = azurerm_container_group.serverless_aci_template.id
} 