module "conf" {
    source = "../module/gcp/conf"
}

# terraform {
#     backend "gcs" {
#         bucket = var.gcp_runner.bucket
#         prefix = var.gcp_runner.prefix
#     }
# }

module "gcp_runner_batch_job_module" {
  count = var.gcp_runner_dind == "true" ? 1 : 0
  depends_on = [ module.conf ]
  source = "../module/gcp/http"
  # source = "../module/gcp/cloudrunservice"
  batch_job = {
    name = var.group_name
    project_id = var.gcp_project
    sa_email = var.gcp_project_sa_email
    api_key = var.gcp_project_apikey
    vpc_name = var.gcp_vpc
    subnet_name = var.gcp_subnet
    tags = {
      product = "serverless-hosted-runner",
      team = "ccoecn",
      maintainer = "hayue2",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
      location = var.gcp_region
    }
  }
  batch_job_container = {
    cpu  = var.runner_cpu
    memory = var.runner_memory
    name = var.gcp_runner.container_name
    image = join("/", ["${var.gcp_region}-docker.pkg.dev", var.gcp_project, "serverless-hosted-runner", var.gcp_runner.container_image])
    image_ver = var.image_ver
    image_retrieve_psw = var.IMAGE_RETRIEVE_PWD
    image_retrieve_uname = var.IMAGE_RETRIEVE_USERNAME
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.gcp_runner.startup_cmd
    ports_port = var.gcp_runner.ports_port 
    container_type = var.container_type
    runner_id = var.runner_id
    runner_token = var.runner_token
    runner_repurl = var.runner_repurl
    runner_repname = var.runner_repname
    runner_action = var.runner_action
    runner_orgname = var.runner_orgname
    runner_orgowner = var.runner_orgowner
    runner_labels = var.runner_labels
    runner_group = var.runner_group
    cloud_pr = var.cloud_pr
    dis_ip = var.dis_ip
  }
}

module "gcp_runner_module" {
  count = var.gcp_runner_dind == "false" ? 1 : 0
  depends_on = [ module.conf ]
  source = "../module/gcp/cloudrunjob"
  # source = "../module/gcp/cloudrunservice"
  gcp_group = {
    name = var.group_name
    project_id = var.gcp_project
    tags = {
      product = "serverless-hosted-runner",
      team = "ccoecn",
      maintainer = "hayue2",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
      location = var.gcp_region
    }
  }
  gcp_container = {
    cpu  = var.runner_cpu
    memory = var.runner_memory
    name = var.gcp_runner.container_name
    image = join("/", ["${var.gcp_region}-docker.pkg.dev", var.gcp_project, "serverless-hosted-runner", var.gcp_runner.container_image])
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.gcp_runner.startup_cmd
    ports_port = var.gcp_runner.ports_port 
    container_type = var.container_type
    runner_id = var.runner_id
    runner_token = var.runner_token
    runner_repurl = var.runner_repurl
    runner_repname = var.runner_repname
    runner_action = var.runner_action
    runner_orgname = var.runner_orgname
    runner_orgowner = var.runner_orgowner
    runner_labels = var.runner_labels
    runner_group = var.runner_group
    cloud_pr = var.cloud_pr
    dis_ip = var.dis_ip
  }
}