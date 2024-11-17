module "conf" {
    source = "../module/conf" 
}

module "mns_runner" {
  depends_on = [ module.conf ]
  source = "../module/mns"
  mns_vars = {
    name = var.eci_agent.mns_runner_name
    message_retention_period = var.eci_agent.message_retention_period
  }
}

module "mns_pool" {
  depends_on = [ module.conf ]
  source = "../module/mns"
  mns_vars = {
    name = var.eci_agent.mns_pool_name
  }
}