locals {  
    base_cmds = var.gcp_container.startup_cmd == "" ? [] : ["${var.gcp_container.startup_cmd}"]  
    cmds = var.gcp_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.gcp_container.image_ver, "-r", var.gcp_container.runner_lazy_regs, "-a", var.gcp_container.runner_allen_regs, 
        "-m", var.gcp_container.ctx_log_level, "-c", var.gcp_container.cloud_pr, "-t",var.gcp_container.tf_ctl]) : concat(local.base_cmds, 
        ["-t", var.gcp_container.container_type, "-i", var.gcp_container.runner_id, "-k", var.gcp_container.runner_token, 
        "-l", var.gcp_container.runner_repurl, "-n", var.gcp_container.runner_repname, "-a", var.gcp_container.runner_action, 
        "-o", var.gcp_container.runner_orgname, "-p", var.gcp_container.runner_orgowner, "-v", var.gcp_container.image_ver,
        "-b", var.gcp_container.runner_labels, "-g", var.gcp_container.runner_group, "-m", var.gcp_container.ctx_log_level, 
        "-c", var.gcp_container.cloud_pr, "-d",var.gcp_container.dis_ip]) 
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