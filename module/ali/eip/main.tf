# eip module template
resource "alicloud_eip_address" "eip_template" { 
  address_name     = var.eip_var.name
  isp              = var.eip_var.eip_isp 
  netmode          = var.eip_var.eip_netmode
  bandwidth        = var.eip_var.eip_bandwidth
  payment_type     = var.eip_var.eip_payment
} 
output "net_eip" {
  value = alicloud_eip_address.eip_template.id
}
output "net_eip_ipaddress" {
  value = alicloud_eip_address.eip_template.ip_address
}