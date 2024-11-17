variable "slb_var" {
  description = "slb variables"
  type = object({   
    slb_name = optional(string, "serverless_hosted_runner") 
    slb_fe_test_port = optional(string, "443") 
    slb_fe_port = optional(string, "61201") 
    slb_be_port = optional(string, "61201") 
    slb_protocol = optional(string, "http") 
    slb_acl_ipv4_name = optional(string, "serverless_hosted_runner_acl_ipv4") 
    slb_acl_ipv6_name = optional(string, "serverless_hosted_runner_acl_ipv6") 
    vswitch_id = string 
    eip_isp = optional(string, "BGP")
    eip_netmode = optional(string, "public")
    eip_bandwidth = optional(string, "10")
    eip_payment = optional(string, "PayAsYouGo")
  })
  sensitive = false
  nullable  = false
}