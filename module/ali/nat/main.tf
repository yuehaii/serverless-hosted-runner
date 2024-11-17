#nat with eip binded
resource "alicloud_nat_gateway" "nat_gateway_template" {
  vpc_id           = var.net_var.vpc_id
  vswitch_id       = var.net_var.vswitch_id
  nat_type         = var.net_var.nat_type
  payment_type     = var.net_var.nat_payment
  network_type     = var.net_var.nat_network
  nat_gateway_name = var.net_var.net_name
}
resource "alicloud_eip_address" "eip_template" { 
  isp              = var.net_var.eip_isp 
  netmode          = var.net_var.eip_netmode
  bandwidth        = var.net_var.eip_bandwidth
  payment_type     = var.net_var.eip_payment
  address_name     = var.net_var.net_name
}
resource "alicloud_eip_association" "eip_binding_template" {
  allocation_id = alicloud_eip_address.eip_template.id
  instance_id   = alicloud_nat_gateway.nat_gateway_template.id
}
resource "alicloud_snat_entry" "snat_entry_template" {
  # count      = length(var.net_var.vswitch_ids) > 0 ? 0 : length(var.net_var.vswitch_cidrs) 
  snat_table_id     = alicloud_nat_gateway.nat_gateway_template.snat_table_ids
  source_vswitch_id = var.net_var.vswitch_id
  snat_ip           = alicloud_eip_address.eip_template.ip_address
  snat_entry_name   = var.net_var.net_name
}
output "net_gateway" {
  value = alicloud_nat_gateway.nat_gateway_template.id
}