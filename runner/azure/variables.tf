variable "aci_runner" {
  description = "eci runner config"
  type = object({  
    security_group_name = optional(string) 
    container_name      = string
    container_image     = string 
    working_dir         = string
    startup_cmd         = string 
    ports_port          = string 
    restart_policy      = string
  })
}   
variable "gcp_project" { 
    type=string 
    default="none" 
} 
variable "gcp_region"  { 
    type=string 
    default="none" 
}
variable "IMAGE_RETRIEVE_SERVER" { 
    type=string 
    default="none" 
} 
variable "IMAGE_RETRIEVE_USERNAME"  { 
    type=string 
    default="none" 
} 
variable "IMAGE_RETRIEVE_PWD"  { 
    type=string 
    default="none" 
}  
variable "subnet_ids"  { 
    type=string 
    default="none" 
}  
variable "resource_group_name"  { 
    type=string 
    default="none" 
} 
variable "container_type"  { 
    type=string 
    default="none" 
}  
variable "group_name"  { 
    type=string 
    default="serverless-hosted-runner" 
} 
variable "runner_id"  { 
    type=string 
    default="none" 
}  
variable "runner_token"  { 
    type=string 
    default="none" 
}  
variable "runner_repurl"  { 
    type=string 
    default="none" 
} 
variable "runner_repname"  { 
    type=string 
    default="none" 
} 
variable "runner_orgname"  { 
    type=string 
    default="none" 
} 
variable "runner_orgowner"  { 
    type=string 
    default="none" 
} 
variable "runner_action"  { 
    type=string 
    default="none" 
}  
variable "image_ver"  { 
    type=string 
    default="none" 
}  
variable "ctx_log_level"  { 
    type=string 
    default="13" 
}
variable "network_mode"  { 
    type=string 
    default="dynamic" # fixed
}
variable "security_group_id"  { 
    type=string 
    default="" 
}  
variable "vswitch_id"  { 
    type=string 
    default="" 
}   
variable "runner_cpu"  { 
    type=string 
    default="1.0" 
}  
variable "runner_memory"  { 
    type=string 
    default="2.0" 
}
variable "runner_labels"  { 
    type=string 
    default="none" 
}
variable "runner_group"  { 
    type=string 
    default="default" 
}
variable "add_host_ip"  { 
    type=string 
    default="127.0.0.1" 
}
variable "add_host_fqdn"  { 
    type=string 
    default="localhost" 
}
variable "charge_labels"  { 
    type=string 
    default="none" 
}
variable "dns_name_servers"  { 
    type=list(string) 
    default=["10.82.31.69","10.82.31.116"]
}
variable "dns_searches"  { 
    type=list(string) 
    default=["docker.com","googleapis.com","google.com"]
}
variable "workspace_id"  { 
    type=string 
    default="none" 
}
variable "workspace_key"  { 
    type=string 
    default="none" 
}