variable mns_vars {
    description = "mns vars definition"
    type = object ({
        name = string
        delay_seconds = optional(string, "0")
        maximum_message_size = optional(string, "65536")
        message_retention_period = optional(string, "604800")
        visibility_timeout = optional(string, "30")
        polling_wait_seconds = optional(string, "3")
        logging_enabled = optional(string, "true")
    })
}