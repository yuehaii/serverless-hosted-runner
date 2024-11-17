variable eci_image_cache {
    description = "eci image cache paras"
    type = object ({
        name = optional(string, "serverless-runner-icache") 
        eip_isp = optional(string, "BGP")
        eip_netmode = optional(string, "public")
        eip_bandwidth = optional(string, "50")
        eip_payment = optional(string, "PayAsYouGo")
        security_group_id = string
        vswitch_id = string
        image = string
        auth_server = optional(string, "artifactory.cloud.ingka-system.cn")
        auth_user = string
        auth_password = string
    })
}
