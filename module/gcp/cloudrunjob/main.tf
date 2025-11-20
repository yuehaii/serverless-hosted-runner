locals {  
    base_cmds = var.gcp_container.startup_cmd == "" ? [] : ["${var.gcp_container.startup_cmd}"]  
    cmds = var.gcp_container.runner_action == "none" ? concat(local.base_cmds, 
        ["dispatcher", "-v", var.gcp_container.image_ver, "-r", var.gcp_container.runner_lazy_regs, "-a", var.gcp_container.runner_allen_regs, 
        "-m", var.gcp_container.ctx_log_level, "-c", var.gcp_container.cloud_pr, "-t",var.gcp_container.tf_ctl]) : concat(local.base_cmds, 
        ["runner", "-t", var.gcp_container.container_type, "-i", var.gcp_container.runner_id, "-k", var.gcp_container.runner_token, 
        "-l", var.gcp_container.runner_repurl, "-n", var.gcp_container.runner_repname, "-a", var.gcp_container.runner_action, 
        "-o", var.gcp_container.runner_orgname, "-p", var.gcp_container.runner_orgowner, "-v", var.gcp_container.image_ver,
        "-b", var.gcp_container.runner_labels, "-g", var.gcp_container.runner_group, "-m", var.gcp_container.ctx_log_level, 
        "-c", var.gcp_container.cloud_pr, "-d",var.gcp_container.dis_ip, "-r",var.gcp_container.repo_reg_tk]) 
}

resource "random_id" "job_name_suffix" {
  byte_length = 6
}

resource "google_cloud_run_v2_job" "gcp_serverless_runner" {
  provider            = google-beta
  name                = format("%s-%s",var.gcp_group.name, random_id.job_name_suffix.hex)
  project             = var.gcp_group.project_id
  location            = var.gcp_group.location
  deletion_protection = var.gcp_group.deletion_protection
  start_execution_token = "start-once-created"
  template {
    template {
      timeout = "6000s"
      containers {
        name   = var.gcp_container.name
        image = join(":", [var.gcp_container.image, var.gcp_container.image_ver])
        command = local.cmds
        ## TODO if command can't be added with paras, need to use args instead
        # args = [""]
        env {
          name = var.gcp_container.environment_variables_name
          value = var.gcp_container.environment_variables_value
        } 

        env {
          name = var.gcp_container_env_keys.ctx_username
          value = var.gcp_container_env_vals.ctx_username_val
        } 
        env {
          name = var.gcp_container_env_keys.ctx_pwd
          value = var.gcp_container_env_vals.ctx_pwd_val
        } 
        env {
          name = var.gcp_container_env_keys.var_ctx_username
          value = var.gcp_container_env_vals.ctx_username_val
        } 
        env {
          name = var.gcp_container_env_keys.var_ctx_pwd
          value = var.gcp_container_env_vals.ctx_pwd_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_endpoint
          value = var.gcp_container_env_vals.kafka_endpoint_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_topic
          value = var.gcp_container_env_vals.kafka_topic_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_consumer
          value = var.gcp_container_env_vals.kafka_consumer_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_username
          value = var.gcp_container_env_vals.kafka_username_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_pwd
          value = var.gcp_container_env_vals.kafka_pwd_val
        } 
        env {
          name = var.gcp_container_env_keys.kafka_ca
          value = var.gcp_container_env_vals.kafka_ca_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_host
          value = var.gcp_container_env_vals.allan_db_host_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_port
          value = var.gcp_container_env_vals.allan_db_port_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_usr
          value = var.gcp_container_env_vals.allan_db_usr_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_pwd
          value = var.gcp_container_env_vals.allan_db_pwd_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_dbname
          value = var.gcp_container_env_vals.allan_db_dbname_val
        } 
        env {
          name = var.gcp_container_env_keys.allan_db_table
          value = var.gcp_container_env_vals.allan_db_table_val
        } 
        env {
          name = var.gcp_container_env_keys.git_ent_tk
          value = var.gcp_container_env_vals.git_ent_tk_val
        } 
        env {
          name = var.gcp_container_env_keys.git_hub_tk
          value = var.gcp_container_env_vals.git_hub_tk_val
        } 
        env {
          name = var.gcp_container_env_keys.enc_key
          value = var.gcp_container_env_vals.enc_key_val
        } 
        env {
          name = var.gcp_container_env_keys.var_enc_key
          value = var.gcp_container_env_vals.enc_key_val
        } 
        env {
          name = var.gcp_container_env_keys.azure_acr_server
          value = var.gcp_container_env_vals.azure_acr_server_val
        } 
        env {
          name = var.gcp_container_env_keys.azure_acr_username
          value = var.gcp_container_env_vals.azure_acr_username_val
        } 
        env {
          name = var.gcp_container_env_keys.azure_acr_pwd
          value = var.gcp_container_env_vals.azure_acr_pwd_val
        } 
        env {
          name = var.gcp_container_env_keys.var_azure_acr_server
          value = var.gcp_container_env_vals.azure_acr_server_val
        } 
        env {
          name = var.gcp_container_env_keys.var_azure_acr_username
          value = var.gcp_container_env_vals.azure_acr_username_val
        } 
        env {
          name = var.gcp_container_env_keys.var_azure_acr_pwd
          value = var.gcp_container_env_vals.azure_acr_pwd_val
        } 

        ports {
            name = var.gcp_container.ports_name
            container_port = var.gcp_container.ports_port
        }
        resources {
          limits = {
            cpu    = var.gcp_container.cpu
            memory = format("%sGi",var.gcp_container.memory)
          }
        }
      }
    }
  }

}