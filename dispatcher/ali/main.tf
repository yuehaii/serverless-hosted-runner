module "conf" {
    source = "../../module/ali/conf"
}
terraform {  
  backend "oss" { 
    prefix              = "serverless/tfstate" # file path
    region              = "cn-shanghai" 
    tablestore_table    = "serverless_tfstate"
    encrypt = true
  }
} 

module "eci_net" {
    depends_on = [ module.conf ]
    source = "../../module/ali/net"
    net_var = {
        sg_name = var.eci_dispatcher.security_group_name
    }
    count = var.network_mode == "dynamic" ? 1 : 0
}

module "eci_nat" {
    depends_on = [ module.conf, module.eci_net ]
    source = "../../module/ali/nat"
    net_var = {
        vswitch_id = module.eci_net[0].net_vswitch_id
        vpc_id = module.eci_net[0].net_vpc_id 
    }
    count = var.network_mode == "dynamic" ? 1 : 0
} 

module "eci_dispatcher_module" {
  depends_on = [ module.eci_net, module.eci_nat ]
  source = "../../module/ali/eci"
  eci_group = {
    name = var.eci_dispatcher.group_name
    security_group_id = var.security_group_id == "none" ? module.eci_net[0].net_sg_id : var.security_group_id
    vswitch_id = var.vswitch_id == "none" ? module.eci_net[0].net_vswitch_id : var.vswitch_id 
    image_retrieve_server = var.IMAGE_RETRIEVE_SERVER
    image_retrieve_uname = var.IMAGE_RETRIEVE_USERNAME
    image_retrieve_psw = var.IMAGE_RETRIEVE_PWD
    cpu  = var.dispacher_cpu
    memory = var.dispacher_memory
    image_cache = true
    tags = {
      product = "serverless-hosted-runner",
      team = var.team,
      maintainer = "hayue2",
      "k8s.aliyun.com/eci-docker-build-enable" = "true",
      organization = var.runner_orgowner,
      repository = var.runner_repname,
      charge_labels = var.charge_labels
    }
  }
  eci_container = {
    name = var.eci_dispatcher.container_name
    image = var.eci_dispatcher.container_image
    image_ver = var.image_ver
    ctx_log_level = var.ctx_log_level
    working_dir = var.eci_dispatcher.working_dir
    startup_cmd = var.eci_dispatcher.startup_cmd
    ports_port = var.eci_dispatcher.ports_port 
    container_type = var.container_type
    need_privileged = true
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
  eci_mount = var.oss_mount == ""? {} : {
    "oss-sls" =  {
      oss_mount_path = "/go/bin/_work"
      oss_volume_name = var.runner_repname
      oss_mount_path = "/go/bin/_work"
      oss_bucket = var.oss_mount
      oss_url = "oss-cn-shanghai.aliyuncs.com"
      oss_path =  join("/", ["/sls_mount", var.runner_repname, var.runner_id, var.oss_mount])
      oss_ram_role = "sls-mount-oss"
      oss_type = "FlexVolume"
      oss_driver = "alicloud/oss"
    }
  }
  eci_init_container = {}
}

# slb for webhook
module "eci_slb" {
    depends_on = [ module.eci_dispatcher_module ]
    source = "../../module/ali/slb/"
    slb_var = {
        vswitch_id = module.eci_net[0].net_vswitch_id 
        # backend_server_id = module.eci_dispatcher_module.eci_id  
        slb_fe_port = var.eci_dispatcher.ports_port
        slb_be_port = var.eci_dispatcher.ports_port
        slb_protocol = var.eci_dispatcher.protocol
    }
    count = var.network_mode == "dynamic" ? 1 : 0
} 

resource "alicloud_slb_backend_server" "slb_backend_server_template" {
    depends_on = [ module.eci_dispatcher_module, module.eci_slb ] 
    load_balancer_id = var.slb_id == "none" ? module.eci_slb[0].net_slb : var.slb_id
    backend_servers {
      server_id = module.eci_dispatcher_module.eci_id 
      weight    = 100
      type      = "eci"
    }
    count = ((var.slb_id == "" || var.slb_id == "none") && length(module.eci_slb) < 1) ? 0 : 1 
} 