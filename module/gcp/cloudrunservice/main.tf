locals {  
    base_cmds = var.gcp_container.startup_cmd == "" ? [] : ["${var.gcp_container.startup_cmd}"]  
    cmds = var.gcp_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.gcp_container.image_ver, "-r", var.gcp_container.runner_lazy_regs, "-a", var.gcp_container.runner_allen_regs, 
        "-m", var.gcp_container.ctx_log_level, "-c", var.gcp_container.cloud_pr]) : concat(local.base_cmds, 
        ["-t", var.gcp_container.container_type, "-i", var.gcp_container.runner_id, "-k", var.gcp_container.runner_token, 
        "-l", var.gcp_container.runner_repurl, "-n", var.gcp_container.runner_repname, "-a", var.gcp_container.runner_action, 
        "-o", var.gcp_container.runner_orgname, "-p", var.gcp_container.runner_orgowner, "-v", var.gcp_container.image_ver,
        "-b", var.gcp_container.runner_labels, "-g", var.gcp_container.runner_group, "-m", var.gcp_container.ctx_log_level]) 
    liveness_probe_period_seconds        = "10"
    liveness_probe_initial_delay_seconds = "5"
    liveness_probe_success_threshold     = "1"
    liveness_probe_failure_threshold     = "1000"
    liveness_probe_timeout_seconds       = "8"
    liveness_probe_cmds                  = ["pwd"]  
    liveness_probe_port                  = 443 

    readiness_probe_period_seconds        = "10"
    readiness_probe_initial_delay_seconds = "5"
    readiness_probe_success_threshold     = "1"
    readiness_probe_failure_threshold     = "1000"
    readiness_probe_timeout_seconds       = "8"
    readiness_probe_cmds                  = ["pwd"] 
    readiness_probe_port                  = 443  
}

resource "random_id" "runner_name_suffix" {
  byte_length = 6
}

resource "google_cloud_run_v2_service" "gcp_serverless_service" {
  name                = format("%s-%s",var.gcp_group.name, random_id.runner_name_suffix.hex)
  location            = var.gcp_group.location
  deletion_protection = var.gcp_group.deletion_protection
  ingress = var.gcp_group.ingress
  traffic {
    type = var.gcp_group.traffic_type
    percent = var.gcp_group.traffic_percent
  }
  template {
    scaling {
      min_instance_count = var.gcp_group.min_instance_count
      max_instance_count = var.gcp_group.max_instance_count
    }
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
      # startup_probe {
      #   period_seconds        = local.readiness_probe_period_seconds
      #   initial_delay_seconds = local.readiness_probe_initial_delay_seconds
      #   failure_threshold     = local.readiness_probe_failure_threshold
      #   timeout_seconds       = local.readiness_probe_timeout_seconds  
      #   http_get {
      #     path = "/"
      #   }
      # }
      # liveness_probe {
      #   period_seconds        = local.liveness_probe_period_seconds
      #   initial_delay_seconds = local.liveness_probe_initial_delay_seconds 
      #   failure_threshold     = local.liveness_probe_failure_threshold
      #   timeout_seconds       = local.liveness_probe_timeout_seconds
      #   http_get {
      #     path = "/"
      #   }
      # }
    }
  }

}