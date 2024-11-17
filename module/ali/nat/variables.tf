variable "net_var" {
  description = "nat variables"
  type = object({
    # net 
    net_name = optional(string, "serverless_hosted_runner")  
    # vswitch
    vswitch_id  = string         
    # vpc
    vpc_id   = string
    # nat 
    snat_ip  = optional(string)
    nat_type = optional(string, "Enhanced")
    nat_payment = optional(string, "PayAsYouGo")
    nat_network = optional(string, "internet")
    # eip 
    eip_isp = optional(string, "BGP")
    eip_netmode = optional(string, "public")
    eip_bandwidth = optional(string, "50")
    eip_payment = optional(string, "PayAsYouGo")
  })
  sensitive = false
  nullable  = false
}