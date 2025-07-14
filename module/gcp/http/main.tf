locals {  
    base_cmds = [] 
    cmds = var.batch_job_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.batch_job_container.image_ver, "-r", var.batch_job_container.runner_lazy_regs, "-a", var.batch_job_container.runner_allen_regs, 
        "-m", var.batch_job_container.ctx_log_level, "-c", var.batch_job_container.cloud_pr]) : concat(local.base_cmds, 
        ["-t", var.batch_job_container.container_type, "-i", var.batch_job_container.runner_id, "-k", var.batch_job_container.runner_token, 
        "-l", var.batch_job_container.runner_repurl, "-n", var.batch_job_container.runner_repname, "-a", var.batch_job_container.runner_action, 
        "-o", var.batch_job_container.runner_orgname, "-p", var.batch_job_container.runner_orgowner, "-v", var.batch_job_container.image_ver,
        "-b", var.batch_job_container.runner_labels, "-g", var.batch_job_container.runner_group, "-m", var.batch_job_container.ctx_log_level]) 
}

resource "random_id" "job_name_suffix" {
  byte_length = 6
}

data "http" "batch_job" {
  url    = "${var.batch_job.rest_prefix}/projects/${var.batch_job.project_id}/locations/${var.batch_job.location}/jobs"
  method = "POST"
  request_headers = {
    Accept = "application/json"
    Authorization = "Bearer ${var.batch_job.api_key}"
  }
  request_body = <<EOT
  {
    "name": "${format("%s-%s",var.batch_job.name, random_id.job_name_suffix.hex)}",
    "taskGroups":[
      {
        "name": "${format("%s-%s",var.batch_job.name, random_id.job_name_suffix.hex)}",
        "taskSpec": {
          "runnables":{
            "container": {
              "imageUri": "${var.batch_job_container.image}:${var.batch_job_container.image_ver}",
              "entrypoint": "${var.batch_job_container.startup_cmd}",
              "commands": [
                "-t",
                "${var.batch_job_container.container_type}",
                "-i",
                "${var.batch_job_container.runner_id}",
                "-k",
                "${var.batch_job_container.runner_token}",
                "-l",
                "${var.batch_job_container.runner_repurl}",
                "-n",
                "${var.batch_job_container.runner_repname}",
                "-a",
                "${var.batch_job_container.runner_action}", 
                "-o",
                "${var.batch_job_container.runner_orgname}",
                "-p",
                "${var.batch_job_container.runner_orgowner}",
                "-v",
                "${var.batch_job_container.image_ver}",
                "-b",
                "${var.batch_job_container.runner_labels}",
                "-g",
                "${var.batch_job_container.runner_group}",
                "-m",
                "${var.batch_job_container.ctx_log_level}"
                "-c",
                "${var.batch_job_container.cloud_pr}"
                "-d",
                "${var.batch_job_container.dis_ip}"
              ],
              "options": "${var.batch_job_container.optional_paras}",
              "blockExternalNetwork": false,
              "username": "${var.batch_job_container.image_retrieve_uname}",
              "password": "${var.batch_job_container.image_retrieve_psw}",
              "enableImageStreaming": false 
            }
          },
          "maxRetryCount": "0",
          "maxRunDuration": "${var.batch_job_container.max_run_duration}",
          "computeResource": {
            "cpuMilli": "${format("%s000",trim(var.batch_job_container.cpu, " "))}",
            "memoryMib": "${format("%s000",trim(var.batch_job_container.memory, " "))}",
            "bootDiskMib": "${var.batch_job_container.extra_disk}"
          },
          "environment": {
            "variables": {
              "${var.batch_job_container.environment_variables_name}": "${var.batch_job_container.environment_variables_value}"
            }
          },
        },
        "runAsNonRoot": false
      }
    ],
    "allocationPolicy": {
      "serviceAccount": {
        "email": "${var.batch_job.sa_email}"
      },
      "network": {
        "networkInterfaces": [
          {
            "network": "projects/${var.batch_job.project_id}/global/networks/${var.batch_job.vpc_name}",
            "subnetwork": "projects/${var.batch_job.project_id}/regions/${var.batch_job.location}/subnetworks/${var.batch_job.subnet_name}",
            "noExternalIpAddress": false
          }
        ]
      }
    },
    "labels": {
      "product": "${var.batch_job.tags["product"]}",
      "team": "${var.batch_job.tags["team"]}",
      "maintainer": "${var.batch_job.tags["maintainer"]}",
      "organization": "${var.batch_job.tags["organization"]}",
      "repository": "${var.batch_job.tags["repository"]}",
      "charge_labels": "${var.batch_job.tags["charge_labels"]}"
    },
    "logsPolicy": {
      "destination": "CLOUD_LOGGING"
    }
  }
  EOT
}
resource "local_file" "http_response_code" {
    content  = data.http.batch_job.status_code
    filename = "http_response_code.log"
}
resource "local_file" "http_response_body" {
    content  = data.http.batch_job.response_body
    filename = "http_response_body.log"
} 