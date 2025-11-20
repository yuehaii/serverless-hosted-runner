variable "aci_dispatcher" {
  description = "aci dispatcher config"
  type = object({ 
    group_name          = string
    container_name      = string
    container_image     = string
    working_dir         = string
    startup_cmd         = string 
    ports_port          = string 
    protocol            = string 
    org_name            = optional(string)
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
variable "CTX_USERNAME"  { 
    type=string 
    default="none" 
} 
variable "CTX_PWD"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_ENDPOINT"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_TOPIC"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_CONSUMER"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_USERNAME"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_PWD"  { 
    type=string 
    default="none" 
} 
variable "KAFKA_INS_CA_CERT"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_HOST"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_PORT"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_USR"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_PWD"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_DBNAME"  { 
    type=string 
    default="none" 
} 
variable "ALLEN_DB_TABLE"  { 
    type=string 
    default="none" 
}
variable "SLS_GITENT_TK"  { 
    type=string 
    default="none" 
} 
variable "SLS_GITHUB_TK"  { 
    type=string 
    default="none" 
}
variable "SLS_ENC_KEY"  { 
    type=string 
    default="none" 
}
variable "AZURE_ACR_SERVER"  { 
    type=string 
    default="none" 
}
variable "AZURE_ACR_USRNAME"  { 
    type=string 
    default="none" 
}
variable "AZURE_ACR_PWD"  { 
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
    default="serverless-hosted-dispatcher" 
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
    default="none" 
}  
variable "vswitch_id"  { 
    type=string 
    default="none" 
}   
variable "slb_id"  { 
    type=string 
    default="none" 
}   
variable "lazy_regs"  { 
    type=string 
    default="none" 
}
variable "allen_regs"  { 
    type=string 
    default="none" 
}
variable "team"  { 
    type=string 
    default="ccoecn" 
}
variable "charge_labels"  { 
    type=string 
    default="none" 
}
variable "cloud_pr"  { 
    type=string 
    default="ali" 
}
variable "workspace_id"  { 
    type=string 
    default="none" 
}
variable "workspace_key"  { 
    type=string 
    default="none" 
}
variable "dispacher_cpu"  { 
    type=string 
    default="1.0" 
}
variable "dispacher_memory"  { 
    type=string 
    default="2.0" 
}
variable "tf_ctl" {
    type = string
    default = "cmd"
}
variable "oss_mount" {
    type = string
    default = ""
}