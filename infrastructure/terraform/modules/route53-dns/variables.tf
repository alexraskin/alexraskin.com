variable "zone_name" {
  type = string
}

variable "hostnames" {
  type = list(string)
}

variable "target" {
  type = object({
    domain_name    = string
    hosted_zone_id = string
  })
  description = "CloudFront distribution every hostname aliases."
}

variable "acm_validation_options" {
  type = list(object({
    domain_name           = string
    resource_record_name  = string
    resource_record_type  = string
    resource_record_value = string
  }))
}
