# slb tf module template
resource "alicloud_slb_load_balancer" "load_balancer_template" { 
  load_balancer_name = var.slb_var.slb_name
  address_type       = "intranet"
  load_balancer_spec = "slb.s2.small"
  vswitch_id         = var.slb_var.vswitch_id
  tags = {
    info = "create for dispatcher"
  }
  instance_charge_type = "PayBySpec"
}

resource "alicloud_slb_listener" "slb_listener_template" {
  depends_on = [ alicloud_slb_load_balancer.load_balancer_template, alicloud_slb_acl.slb_acl_ipv4_template ]
  load_balancer_id          = alicloud_slb_load_balancer.load_balancer_template.id
  backend_port              = var.slb_var.slb_be_port
  frontend_port             = var.slb_var.slb_fe_port
  protocol                  = var.slb_var.slb_protocol
  bandwidth                 = 10
  sticky_session            = "off"  
  health_check              = "off" 
  health_check_uri          = "/" 
  healthy_threshold         = 3
  unhealthy_threshold       = 3
  health_check_timeout      = 5
  health_check_interval     = 2
  health_check_http_code    = "http_2xx,http_3xx" 
  acl_status                = "off"
  acl_type                  = "white"
  acl_id                    = alicloud_slb_acl.slb_acl_ipv4_template.id
  request_timeout           = 60
  idle_timeout              = 15
} 
resource "alicloud_slb_listener" "slb_listener_local" {
  depends_on = [ alicloud_slb_load_balancer.load_balancer_template, alicloud_slb_acl.slb_acl_ipv4_template ]
  load_balancer_id          = alicloud_slb_load_balancer.load_balancer_template.id
  backend_port              = var.slb_var.slb_be_port
  frontend_port             = var.slb_var.slb_fe_test_port
  protocol                  = var.slb_var.slb_protocol
  bandwidth                 = 10
  sticky_session            = "off"  
  health_check              = "off" 
  health_check_uri          = "/" 
  healthy_threshold         = 3
  unhealthy_threshold       = 3
  health_check_timeout      = 5
  health_check_interval     = 2
  health_check_http_code    = "http_2xx,http_3xx" 
  acl_status                = "off"
  acl_type                  = "white"
  acl_id                    = alicloud_slb_acl.slb_acl_ipv4_template.id
  # acl_id                    = join(",", [alicloud_slb_acl.slb_acl_ipv4_template.id, alicloud_slb_acl.slb_acl_ipv6_template.id])
  request_timeout           = 60
  idle_timeout              = 15
}  
resource "alicloud_slb_acl" "slb_acl_ipv4_template" { 
  name       = var.slb_var.slb_acl_ipv4_name
  ip_version = "ipv4"
} 

resource "alicloud_eip_address" "slb_eip_template" { 
  isp              = var.slb_var.eip_isp 
  netmode          = var.slb_var.eip_netmode
  bandwidth        = var.slb_var.eip_bandwidth
  payment_type     = var.slb_var.eip_payment
} 
resource "alicloud_eip_association" "slb_eip_binding_template" {
  depends_on = [ alicloud_eip_address.slb_eip_template, alicloud_slb_load_balancer.load_balancer_template ]
  allocation_id = alicloud_eip_address.slb_eip_template.id
  instance_id   = alicloud_slb_load_balancer.load_balancer_template.id
}
output "net_slb" {
  value = alicloud_slb_load_balancer.load_balancer_template.id
}