variable "eip_var" {
  description = "eip variables"
  type = object({   
    name    = optional(string, "serverles-eip")
    eip_isp = optional(string, "BGP")
    eip_netmode = optional(string, "public")
    eip_bandwidth = optional(string, "50")
    eip_payment = optional(string, "PayAsYouGo")
  })
  sensitive = false
  nullable  = false
}