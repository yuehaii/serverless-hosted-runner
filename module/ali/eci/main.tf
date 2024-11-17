## eci tf module template
locals {  
    base_cmds = var.eci_container.startup_cmd == "" ? [] : ["${var.eci_container.startup_cmd}"]  
    cmds = var.eci_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.eci_container.image_ver, "-r", var.eci_container.runner_lazy_regs, "-a", var.eci_container.runner_allen_regs, 
        "-m", var.eci_container.ctx_log_level, "-c", var.eci_container.cloud_pr]) : concat(local.base_cmds, 
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
    containers {
        name = var.eci_container.name
        image = join(":", [var.eci_container.image, var.eci_container.image_ver])
        security_context {
            privileged = var.eci_container.need_privileged
        }
        working_dir = var.eci_container.working_dir
        image_pull_policy = var.eci_container.image_pull_policy
        commands = local.cmds
        ports {
            port     = var.eci_container.ports_port
            protocol = var.eci_container.ports_protocol
        }
        environment_vars { 
            key   = var.eci_container.environment_key
            value = var.eci_container.environment_val 
        } 
        # # pass testing. ref link below for oss ram role: 
        # # https://www.alibabacloud.com/help/zh/eci/user-guide/mount-an-oss-bucket-to-an-elastic-container-instance-as-a-volume?spm=a2c63.p38356.0.0.47203166MHLy8b#6b389eb05f549
        # volume_mounts {
        #     name       = var.eci_container.oss_volume_name
        #     mount_path = var.eci_container.oss_mount_path
        #     read_only  = false
        # }
        liveness_probe {
            period_seconds        = local.liveness_probe_period_seconds
            initial_delay_seconds = local.liveness_probe_initial_delay_seconds
            success_threshold     = local.liveness_probe_success_threshold
            failure_threshold     = local.liveness_probe_failure_threshold
            timeout_seconds       = local.liveness_probe_timeout_seconds
            exec {
                commands = local.liveness_probe_cmds
            }
        }
        readiness_probe {
            period_seconds        = local.readiness_probe_period_seconds
            initial_delay_seconds = local.readiness_probe_initial_delay_seconds
            success_threshold     = local.readiness_probe_success_threshold
            failure_threshold     = local.readiness_probe_failure_threshold
            timeout_seconds       = local.readiness_probe_timeout_seconds
            exec {
                commands = local.readiness_probe_cmds
            }
        }
    }
    # # incase we need to store the state after runner reboot
    # volumes {
    #     name = var.eci_container.oss_volume_name
    #     type = "FlexVolume"
    #     flex_volume_driver = "alicloud/oss" 
    #     flex_volume_options = format("{\"bucket\":\"%s\",\"url\":\"%s\",\"path\":\"%s\",\"ramRole\":\"%s\"}",
    #         var.eci_container.oss_bucket, var.eci_container.oss_url, 
    #         var.eci_container.oss_path, var.eci_container.oss_ram_role)
    # }
} 
output "eci_id" {
  value = alicloud_eci_container_group.serverless_eci_template.id
} 