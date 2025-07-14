
    containers {  
        name = "${pool_container_name}"
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
            value = "${container_id}"
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
