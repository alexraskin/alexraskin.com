variable "zone_name" {
  type = string
}

variable "hostnames" {
  type = list(string)
}

variable "target" {
  type        = string
  description = "CNAME target for every hostname (CloudFront domain)."
}

variable "acm_validation_options" {
  type = list(object({
    domain_name           = string
    resource_record_name  = string
    resource_record_type  = string
    resource_record_value = string
  }))
}

variable "cloudflare_comment" {
  type = string
}
