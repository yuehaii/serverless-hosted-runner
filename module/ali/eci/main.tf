## eci tf module template
locals {  
    base_cmds = var.eci_container.startup_cmd == "" ? [] : ["${var.eci_container.startup_cmd}"]  
    cmds = var.eci_container.runner_action == "none" ? concat(local.base_cmds, 
        ["-v", var.eci_container.image_ver, "-r", var.eci_container.runner_lazy_regs, "-a", var.eci_container.runner_allen_regs, 
        "-m", var.eci_container.ctx_log_level, "-c", var.eci_container.cloud_pr, "-t",var.eci_container.tf_ctl]) : concat(local.base_cmds, 
        ["-t", var.eci_container.container_type, "-i", var.eci_container.runner_id, "-k", var.eci_container.runner_token, 
        "-l", var.eci_container.runner_repurl, "-n", var.eci_container.runner_repname, "-a", var.eci_container.runner_action, 
        "-o", var.eci_container.runner_orgname, "-p", var.eci_container.runner_orgowner, "-v", var.eci_container.image_ver,
        "-b", var.eci_container.runner_labels, "-g", var.eci_container.runner_group, "-m", var.eci_container.ctx_log_level,
        "-c", var.eci_container.cloud_pr, "-d",var.eci_container.dis_ip]) 
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
    cpu = trim(var.eci_group.cpu, " ")
    memory = trim(var.eci_group.memory, " ")
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
        dynamic volume_mounts {
            for_each = var.eci_mount
            content {
                name       = volume_mounts.value["oss_volume_name"]
                mount_path = volume_mounts.value["oss_mount_path"]
                read_only  = false
            }
        }
        volume_mounts {
            mount_path = var.eci_container.docker_mount_path
            read_only  = false
            name       = var.eci_container.docker_volume_name
        }
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
    dynamic init_containers {
        for_each = var.eci_init_container
        content {
            name = init_containers.value["init_container_name"]
            image             = init_containers.value["init_container_image"]
            image_pull_policy = init_containers.value["init_container_pullpolicy"]
            commands          = init_containers.value["init_container_cmds"]
            args              = init_containers.value["init_container_args"]
            security_context {
                capability {
                add = [ "CAP_SYS_ADMIN" ]
                }
            }
        }

    }
    dynamic volumes {
        for_each = var.eci_mount
        content {
            name = volumes.value["oss_volume_name"]
            type = volumes.value["oss_type"]
            flex_volume_driver = volumes.value["oss_driver"]
            flex_volume_options = format("{\"bucket\":\"%s\",\"url\":\"%s\",\"path\":\"%s\",\"ramRole\":\"%s\"}",
                volumes.value["oss_bucket"], volumes.value["oss_url"], 
                volumes.value["oss_path"], volumes.value["oss_ram_role"])
        }
    }
    volumes {
        name = var.eci_container.docker_volume_name
        type = var.eci_container.docker_volume_type
    }
} 
output "eci_id" {
  value = alicloud_eci_container_group.serverless_eci_template.id
} 