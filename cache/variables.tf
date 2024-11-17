 variable "vswitch_id" { 
    type=string 
}
variable "security_group_id"  { 
    type=string 
}
variable "image_name"  { 
    type=string 
    default = "artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci"
}
variable "image_ver"  { 
    type=string 
} 
variable "username"  { 
    type=string
}
variable "password"  { 
    type=string 
}