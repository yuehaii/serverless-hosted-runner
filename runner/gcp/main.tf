module "conf" {
    source = "../module/gcp/conf"
}

# terraform {
#     backend "gcs" {
#         bucket = var.gcp_runner.bucket
#         prefix = var.gcp_runner.prefix
#     }
# }

module "gcp_runner_module" {
  depends_on = [ module.conf ]
  source = "../module/gcp/cloudrunjob"
  # source = "../module/gcp/cloudrunservice"
  gcp_group = {
    name = var.group_name
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
    image = join("/", ["gcr.io", var.gcp_project, var.gcp_runner.container_image])
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
  }
}