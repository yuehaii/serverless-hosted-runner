### need pre-create rg, vnet/subnet
# resource "azurerm_resource_group" "serverless_rg_template" {
#   name     = "aci-rg"
#   location = "chinanorth3"
# }

locals {  
    base_cmds = var.aci_container.startup_cmd == "" ? [] : ["${var.aci_container.startup_cmd}"]  
    cmds = var.aci_container.runner_action == "none" ? concat(local.base_cmds, 
        ["dispatcher", "-v", var.aci_container.image_ver, "-r", var.aci_container.runner_lazy_regs, "-a", var.aci_container.runner_allen_regs, 
        "-m", var.aci_container.ctx_log_level, "-c", var.aci_container.cloud_pr, "-t",var.aci_container.tf_ctl]) : concat(local.base_cmds, 
        ["runner", "-t", var.aci_container.container_type, "-i", var.aci_container.runner_id, "-k", var.aci_container.runner_token, 
        "-l", var.aci_container.runner_repurl, "-n", var.aci_container.runner_repname, "-a", var.aci_container.runner_action, 
        "-o", var.aci_container.runner_orgname, "-p", var.aci_container.runner_orgowner, "-v", var.aci_container.image_ver,
        "-b", var.aci_container.runner_labels, "-g", var.aci_container.runner_group, "-m", var.aci_container.ctx_log_level,
        "-c", var.aci_container.cloud_pr, "-d",var.aci_container.dis_ip, "-r",var.aci_container.repo_reg_tk]) 
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

resource "random_id" "job_name_suffix" {
  byte_length = 6
}

resource "azurerm_container_group" "serverless_aci_template" {
  name                = format("%s-%s",var.aci_group.name, random_id.job_name_suffix.hex)
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
  subnet_ids =  var.aci_group.ip_address_type == "Private" ? [var.aci_group.subnet_ids] : []
  # dns_config {
  #   nameservers = var.aci_group.dns_name_servers
  #   search_domains = var.aci_group.dns_searches
  #   options = ["edns0"]
  # }
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
    # TODO: there is a bug in latest pr 4.11. it can't identify number. rollback to 4.3
    # cpu    = tonumber(var.aci_container.cpu)
    # memory = tonumber(var.aci_container.memory)
    commands = local.cmds
    ports {
        port     = var.aci_container.ports_port
        protocol = var.aci_container.ports_protocol
    } 
    security {
        privilege_enabled = var.aci_group.sku == "Confidential" ? true : false
    }
    environment_variables = {
        runner = "aci", 
        TF_VAR_IMAGE_RETRIEVE_SERVER = var.aci_group.image_retrieve_server,
        TF_VAR_IMAGE_RETRIEVE_USERNAME = var.aci_group.image_retrieve_uname,
        TF_VAR_IMAGE_RETRIEVE_PWD = var.aci_group.image_retrieve_psw,
        contextusername = var.aci_container_env_vals.ctx_username_val,
        contextpassword = var.aci_container_env_vals.ctx_pwd_val,
        TF_VAR_CTX_USERNAME = var.aci_container_env_vals.ctx_username_val,
        TF_VAR_CTX_PWD = var.aci_container_env_vals.ctx_pwd_val,
        KAFKA_INS_ENDPOINT = var.aci_container_env_vals.kafka_endpoint_val,
        KAFKA_INS_TOPIC = var.aci_container_env_vals.kafka_topic_val,
        KAFKA_INS_CONSUMER = var.aci_container_env_vals.kafka_consumer_val,
        KAFKA_INS_USERNAME = var.aci_container_env_vals.kafka_username_val,
        KAFKA_INS_PWD = var.aci_container_env_vals.kafka_pwd_val,
        KAFKA_INS_CA_CERT = var.aci_container_env_vals.kafka_ca_val,
        ALLEN_DB_HOST = var.aci_container_env_vals.allan_db_host_val,
        ALLEN_DB_PORT = var.aci_container_env_vals.allan_db_port_val,
        ALLEN_DB_USR = var.aci_container_env_vals.allan_db_usr_val,
        ALLEN_DB_PWD = var.aci_container_env_vals.allan_db_pwd_val,
        ALLEN_DB_DBNAME = var.aci_container_env_vals.allan_db_dbname_val,
        ALLEN_DB_TABLE = var.aci_container_env_vals.allan_db_table_val,
        SLS_GITENT_TK = var.aci_container_env_vals.git_ent_tk_val,
        SLS_GITHUB_TK = var.aci_container_env_vals.git_hub_tk_val,
        SLS_ENC_KEY = var.aci_container_env_vals.enc_key_val,
        TF_VAR_SLS_ENC_KEY = var.aci_container_env_vals.enc_key_val,
        AZURE_ACR_SERVER = var.aci_container_env_vals.azure_acr_server_val,
        AZURE_ACR_USRNAME = var.aci_container_env_vals.azure_acr_username_val,
        AZURE_ACR_PWD = var.aci_container_env_vals.azure_acr_pwd_val,
        TF_VAR_AZURE_ACR_SERVER = var.aci_container_env_vals.azure_acr_server_val,
        TF_VAR_AZURE_ACR_USRNAME = var.aci_container_env_vals.azure_acr_username_val,
        TF_VAR_AZURE_ACR_PWD = var.aci_container_env_vals.azure_acr_pwd_val,
    }
    
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