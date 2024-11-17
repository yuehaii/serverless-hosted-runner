variable "net_var" {
  description = "network variables"
  type = object({
    # net 
    net_name = optional(string, "serverless_hosted_runner") 
    # route
    route_table_id = optional(string, "") 
    # vswitch
    vswitch_ids   = optional(list(string), [])
    vswitch_cidrs = optional(list(string), ["10.1.0.0/16"]) 
    vswitch_cidr_gw = optional(string, "10.3.0.0/16")
    # terway
    terway_vswitch_ids   = optional(list(string), []) 
    terway_vswitch_cidrs = optional(list(string), ["10.4.0.0/16"])
    # vpc
    vpc_id   = optional(string, "")
    vpc_cidr = optional(string, "10.0.0.0/8")
    # sg
    sg_name = string
    sg_id = optional(string, "") 
    sg_type = optional(string, "normal")   
    #sg rule ingress 
    sg_ingress_ip_protocol = optional(string, "all")
    sg_ingress_nic_type = optional(string, "intranet")
    sg_ingress_policy = optional(string, "accept")
    sg_ingress_port_range = optional(string, "-1/-1")
    sg_ingress_priority = optional(string, "1")
    sg_ingress_cidr_ip = optional(string, "0.0.0.0/0") 
    #sg rule egress 
    sg_egress_ip_protocol = optional(string, "all")
    sg_egress_nic_type = optional(string, "intranet")
    sg_egress_policy = optional(string, "accept")
    sg_egress_port_range = optional(string, "-1/-1")
    sg_egress_priority = optional(string, "1")
    sg_egress_cidr_ip = optional(string, "0.0.0.0/0") 

  })
  sensitive = false
  nullable  = false
}