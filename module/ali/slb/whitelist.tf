# ref https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-githubs-ip-addresses
#     https://api.github.com/meta for webhook section
# TODO: those CIDR are not fixed values. Need to find a good workaround to update the whitelist
#       may need to use bash or simple app to fetch those CIDR periodically.
resource "alicloud_slb_acl_entry_attachment" "hookone" {
  depends_on = [ alicloud_slb_acl.slb_acl_ipv4_template ]
  acl_id  = alicloud_slb_acl.slb_acl_ipv4_template.id
  entry   = "192.30.252.0/22"
  comment = "github webhook CIDR#1"
}
resource "alicloud_slb_acl_entry_attachment" "hooktwo" {
  depends_on = [ alicloud_slb_acl.slb_acl_ipv4_template ]
  acl_id  = alicloud_slb_acl.slb_acl_ipv4_template.id
  entry   = "185.199.108.0/22"
  comment = "github webhook CIDR#2"
}
resource "alicloud_slb_acl_entry_attachment" "hookthree" {
  depends_on = [ alicloud_slb_acl.slb_acl_ipv4_template ]
  acl_id  = alicloud_slb_acl.slb_acl_ipv4_template.id
  entry   = "140.82.112.0/20"
  comment = "github webhook CIDR#3"
}
resource "alicloud_slb_acl_entry_attachment" "hookfour" {
  depends_on = [ alicloud_slb_acl.slb_acl_ipv4_template ]
  acl_id  = alicloud_slb_acl.slb_acl_ipv4_template.id
  entry   = "143.55.64.0/20"
  comment = "github webhook CIDR#4"
}

# resource "alicloud_slb_acl_entry_attachment" "hookfive" {
#   acl_id  = alicloud_slb_acl.slb_acl_ipv6_template.id
#   entry   = "2a0a:a440::/29"
#   comment = "github webhook CIDR#5"
# }
# resource "alicloud_slb_acl_entry_attachment" "hooksix" {
#   acl_id  = alicloud_slb_acl.slb_acl_ipv6_template.id
#   entry   = "2606:50c0::/32"
#   comment = "github webhook CIDR#6"
# }