locals {  
    base_cmds = var.gcp_container.startup_cmd == "" ? [] : ["${var.gcp_container.startup_cmd}"]  
    cmds = var.gcp_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.gcp_container.image_ver, "-r", var.gcp_container.runner_lazy_regs, "-a", var.gcp_container.runner_allen_regs, 
        "-m", var.gcp_container.ctx_log_level, "-c", var.gcp_container.cloud_pr]) : concat(local.base_cmds, 
        ["-t", var.gcp_container.container_type, "-i", var.gcp_container.runner_id, "-k", var.gcp_container.runner_token, 
        "-l", var.gcp_container.runner_repurl, "-n", var.gcp_container.runner_repname, "-a", var.gcp_container.runner_action, 
        "-o", var.gcp_container.runner_orgname, "-p", var.gcp_container.runner_orgowner, "-v", var.gcp_container.image_ver,
        "-b", var.gcp_container.runner_labels, "-g", var.gcp_container.runner_group, "-m", var.gcp_container.ctx_log_level]) 
}

resource "random_id" "batch_name_suffix" {
  byte_length = 6
}
 
resource "google_cloud_scheduler_job" "batch_job_runner" {
  paused           = false
  name             = "batch-job-invoker"
  project          = var.batch_group.project_id
  region           = var.batch_group.location
  # schedule         = "*/1 * * * *"
  time_zone        = "America/Los_Angeles"
  attempt_deadline = "180s"

  retry_config {
    max_doublings        = 5
    max_retry_duration   = "0s"
    max_backoff_duration = "3600s"
    min_backoff_duration = "5s"
  }

  http_target {
    http_method = "POST"
    uri = "https://batch.googleapis.com/v1/projects/${var.batch_group.project_number}/locations/${var.batch_group.locationn}/jobs"
    headers = {
      "Content-Type" = "application/json"
      "User-Agent"   = "Google-Cloud-Scheduler"
    }
    body = base64encode(<<EOT
    {
      "taskGroups":[
        {
          "taskSpec": {
            "runnables":{
              "script": {
                "text": "echo Hello world! This job was created using Terraform and Cloud Scheduler."
              }
            }
          }
        }
      ],
      "allocationPolicy": {
        "serviceAccount": {
          "email": "${var.batch_container.batch_service_account_email}"
        }
      },
      "labels": {
        "source": "terraform_and_cloud_scheduler_tutorial"
      },
      "logsPolicy": {
        "destination": "CLOUD_LOGGING"
      }
    }
    EOT
    )
    oauth_token {
      scope                 = "https://www.googleapis.com/auth/cloud-platform"
      service_account_email = var.batch_container.cloud_scheduler_service_account_email
    }
  }
}
