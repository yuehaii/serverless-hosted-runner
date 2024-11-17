## eci image cache

resource "alicloud_eip_address" "default" {
  isp                       = var.eci_image_cache.eip_isp
  address_name              = var.eci_image_cache.name
  netmode                   = var.eci_image_cache.eip_netmode
  bandwidth                 = var.eci_image_cache.eip_bandwidth
  payment_type              = var.eci_image_cache.eip_payment
}

resource "alicloud_eci_image_cache" "default" {
  image_cache_name  = var.eci_image_cache.name
  images            = [var.eci_image_cache.image]
  security_group_id = var.eci_image_cache.security_group_id
  vswitch_id        = var.eci_image_cache.vswitch_id
  eip_instance_id   = alicloud_eip_address.default.id
  image_registry_credential {
    user_name = var.eci_image_cache.auth_user
    password = var.eci_image_cache.auth_password
    server = var.eci_image_cache.auth_server
  }
}