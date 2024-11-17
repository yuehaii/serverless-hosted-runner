variable "eci_agent" {
  description = "eci agent config"
  type = object({ 
    mns_runner_name = string
    mns_pool_name = string
    message_retention_period = optional(string, "3600")  
  })
}