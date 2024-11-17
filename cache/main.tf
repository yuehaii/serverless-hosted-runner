module "conf" {
    source = "../../module/ali/conf"
}

module "eci_icache" {
    depends_on = [ module.conf ]
    source = "../../module/ali/icache"
    eci_image_cache = {
        vswitch_id = var.vswitch_id
        security_group_id = var.security_group_id
        image = "${var.image_name}:${var.image_ver}" 
        auth_user = var.username
        auth_password = var.password 
    }
}