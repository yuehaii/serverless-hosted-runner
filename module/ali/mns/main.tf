# mns module template 

resource "alicloud_message_service_queue" "queue" {
  queue_name               = var.mns_vars.name
  delay_seconds            = var.mns_vars.delay_seconds
  maximum_message_size     = var.mns_vars.maximum_message_size
  message_retention_period = var.mns_vars.message_retention_period
  visibility_timeout       = var.mns_vars.visibility_timeout
  polling_wait_seconds     = var.mns_vars.polling_wait_seconds
  logging_enabled          = var.mns_vars.logging_enabled
}