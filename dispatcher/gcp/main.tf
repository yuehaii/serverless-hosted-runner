module "conf" {
    source = "../../module/gcp/conf"
    gcp_project = var.gcp_project
    gcp_region = var.gcp_region
}  

terraform {
    backend "gcs" {
        bucket = var.gcp_dispatcher.bucket
        prefix = var.gcp_dispatcher.prefix
    }
}

module "gcp_dispatcher_module" {
  depends_on = [ module.conf ]
  source = "../../module/gcp/cloudrunservice"
  gcp_group = {
    name = var.gcp_dispatcher.group_name
    cpu  = var.dispacher_cpu
    memory = var.dispacher_memory
    tags = {
      product = "serverless-hosted-runner",
      team = var.team,
      maintainer = "hayue2",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
      location = var.gcp_region
    }
  }
  gcp_container = {
    name = var.gcp_dispatcher.container_name
    image = join("/", ["${var.gcp_region}-docker.pkg.dev", var.gcp_project, "serverless-hosted-runner", var.gcp_dispatcher.container_image])
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    startup_cmd = var.gcp_dispatcher.startup_cmd
    ports_port = var.gcp_dispatcher.ports_port 
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