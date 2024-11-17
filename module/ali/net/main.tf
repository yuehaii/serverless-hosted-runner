# vswitch tf module template

# az
data "alicloud_enhanced_nat_available_zones" "enhanced" {}

# vpc
resource "alicloud_vpc" "vpc_template" {
  count      = var.net_var.vpc_id == "" ? 1 : 0
  vpc_name   = var.net_var.net_name
  cidr_block = var.net_var.vpc_cidr
}
output "net_vpc_id" {
  value = alicloud_vpc.vpc_template[0].id
}

#route
# resource "alicloud_route_table" "route_template" { 
#   vpc_id           = var.net_var.vpc_id == "" ? join("", alicloud_vpc.vpc_template.*.id) : var.net_var.vpc_id
#   route_table_name = "network outbound"
#   associate_type   = "VSwitch"
# }
# resource "alicloud_route_entry" "route_entry_template" {
#   route_table_id        = var.route_table_id.vpc_id == "" ? join("", alicloud_route_table.route_template.*.id) : var.net_var.route_table_id
#   destination_cidrblock = "0.0.0.0/0"
#   nexthop_type          = "Instance"
#   nexthop_id            = alicloud_instance.foo.id
# }

# vswitch
resource "alicloud_vswitch" "vswitches_template" {
  count        = length(var.net_var.vswitch_ids) > 0 ? 0 : length(var.net_var.vswitch_cidrs)
  vpc_id       = var.net_var.vpc_id == "" ? join("", alicloud_vpc.vpc_template.*.id) : var.net_var.vpc_id
  cidr_block   = element(var.net_var.vswitch_cidrs, count.index)
  zone_id      = data.alicloud_enhanced_nat_available_zones.enhanced.zones[count.index].zone_id
  vswitch_name = var.net_var.net_name
  #zone_id    = data.alicloud_zones.switch_zone.zones[count.index].id
}
output "net_vswitch_id" {
  value = alicloud_vswitch.vswitches_template[0].id
}

# gw 
# resource "alicloud_vpn_gateway" "gateway_template" {  
#   vpc_id     = var.net_var.vpc_id == "" ? join("", alicloud_vpc.vpc_template.*.id) : var.net_var.vpc_id
#   bandwidth  = "10"
#   enable_ssl = true
#   #TODO: PostPaid dose not work. But tf can't delete the PrePaid GW. 
#   #ref: https://registry.terraform.io/providers/aliyun/alicloud/latest/docs/resources/vpn_gateway#instance_charge_type
#   instance_charge_type = "PrePaid" 
#   network_type = "public"
#   vswitch_id = length(var.net_var.vswitch_ids) == 0 ? alicloud_vswitch.vswitches_template[0].id : var.net_var.vswitch_ids[0] 
# }

# terway
# resource "alicloud_vswitch" "terway_vswitches_template" {
#   count      = length(var.net_var.terway_vswitch_ids) > 0 ? 0 : length(var.net_var.terway_vswitch_cidrs)
#   vpc_id     = var.net_var.vpc_id == "" ? join("", alicloud_vpc.vpc_template.*.id) : var.net_var.vpc_id
#   cidr_block = element(var.net_var.terway_vswitch_cidrs, count.index)
#   zone_id    = data.alicloud_enhanced_nat_available_zones.enhanced.zones[count.index].zone_id
# }

# sg
resource "alicloud_security_group" "security_group_template" {
  count      = var.net_var.sg_id == "" ? 1 : 0 
  #name       = var.net_var.sg_name
  vpc_id     = var.net_var.vpc_id == "" ? join("", alicloud_vpc.vpc_template.*.id) : var.net_var.vpc_id
  security_group_type = var.net_var.sg_type 
  name       = var.net_var.net_name
}
resource "alicloud_security_group_rule" "allow_all_ingress" {
  type              = "ingress"
  ip_protocol       = var.net_var.sg_ingress_ip_protocol
  nic_type          = var.net_var.sg_ingress_nic_type
  policy            = var.net_var.sg_ingress_policy
  port_range        = var.net_var.sg_ingress_port_range
  priority          = var.net_var.sg_ingress_priority
  security_group_id = var.net_var.sg_id == "" ? join("", alicloud_security_group.security_group_template[*].id) : var.net_var.sg_id
  cidr_ip           = var.net_var.sg_ingress_cidr_ip
}
resource "alicloud_security_group_rule" "allow_all_egress" {
  type              = "egress"
  ip_protocol       = var.net_var.sg_egress_ip_protocol
  nic_type          = var.net_var.sg_egress_nic_type
  policy            = var.net_var.sg_egress_policy
  port_range        = var.net_var.sg_egress_port_range
  priority          = var.net_var.sg_egress_priority
  security_group_id = var.net_var.sg_id == "" ? join("", alicloud_security_group.security_group_template[*].id) : var.net_var.sg_id
  cidr_ip           = var.net_var.sg_egress_cidr_ip
}
output "net_sg_id" {
  value = alicloud_security_group.security_group_template[0].id
}
