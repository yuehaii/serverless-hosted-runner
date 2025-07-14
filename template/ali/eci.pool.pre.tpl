## eci tf module template
locals {  
    base_cmds = var.eci_container.startup_cmd == "" ? [] : ["${var.eci_container.startup_cmd}"]  
    cmds = var.eci_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.eci_container.image_ver, "-r", var.eci_container.runner_lazy_regs, "-a", var.eci_container.runner_allen_regs, 
        "-m", var.eci_container.ctx_log_level, "-c", var.eci_container.cloud_pr, "-t",var.eci_container.tf_ctl]) : concat(local.base_cmds, 
        ["-t", var.eci_container.container_type, "-i", var.eci_container.runner_id, "-k", var.eci_container.runner_token, 
        "-l", var.eci_container.runner_repurl, "-n", var.eci_container.runner_repname, "-a", var.eci_container.runner_action, 
        "-o", var.eci_container.runner_orgname, "-p", var.eci_container.runner_orgowner, "-v", var.eci_container.image_ver,
        "-b", var.eci_container.runner_labels, "-g", var.eci_container.runner_group, "-m", var.eci_container.ctx_log_level]) 
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

resource "alicloud_eci_container_group" "serverless_eci_template" {
    container_group_name = var.eci_group.name
    cpu = var.eci_group.cpu
    memory = var.eci_group.memory
    auto_match_image_cache = var.eci_group.image_cache
    restart_policy = var.eci_group.restart_policy
    security_group_id =  var.eci_group.security_group_id
    vswitch_id = var.eci_group.vswitch_id
    # security_group_id =  var.eci_group.security_group_id == "" ? join("", module.eci_net.net_sg_ids) : var.eci_group.security_group_id 
    # vswitch_id = var.eci_group.vswitch_id == "" ? join(",", module.eci_net.net_vswitch_ids) : var.eci_group.vswitch_id
    tags = var.eci_group.tags
    host_aliases {
      ip = var.eci_group.add_host_ip
      hostnames = [ var.eci_group.add_host_fqdn ]
    }
    image_registry_credential {
      user_name = var.eci_group.image_retrieve_uname 
      password = var.eci_group.image_retrieve_psw
      server = var.eci_group.image_retrieve_server
    }
    dns_config {
      name_servers = var.eci_group.dns_name_servers
      searches = var.eci_group.dns_searches
      options {
        name = "edns0"
        value = "true"
      }
    }
    