eci_dispatcher = { 
    group_name          = "serverless-dispatcher-group"
    security_group_name = "dispatcher-security-group" 
    container_name      = "dispatcher-container"
    container_image     = "artifactory.cloud.ingka-system.cn/ccoecn-docker-virtual/serverless-hosted-dispatcher" 
    org_name            = "ingka-group-digital"
    ports_port          = "61201"
    startup_cmd         = "./dispatcher"
    working_dir         = "/go/bin/"
    protocol            = "http"
}