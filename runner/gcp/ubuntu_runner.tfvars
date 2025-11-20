gcp_runner = { 
  security_group_name = "gcp-runner-security-group"
  container_name      = "runner-container" 
  cpu                 = "1.0" 
  memory              = "2Gi" 
  container_image     = "serverless-hosted-runner-eci"  
  startup_cmd         = "./serverless-runner"
  working_dir         = "/go/bin"
  ports_port          = "80"
  restart_policy      = "Never"
}