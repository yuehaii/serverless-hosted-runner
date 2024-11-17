aci_runner = { 
  security_group_name = "eci-runner-security-group"
  container_name      = "runner-container" 
  cpu                 = "1.0" 
  memory              = "2.0" 
  container_image     = "artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-runner-eci"  
  startup_cmd         = "./runner"
  working_dir         = "/go/bin"
  ports_port          = "80"
  restart_policy      = "Never"
}