variable "environment" {
  type = string
  validation {
    condition     = contains(["dev", "prod"], var.environment)
    error_message = "environment must be dev or prod."
  }
}

variable "aws_region" {
  description = "Region for the bucket (CloudFront/ACM use the us_east_1 alias)."
  type        = string
}

variable "tags" {
  description = "Extra tags merged into the default tags."
  type        = map(string)
  default     = {}
}

variable "domain_name" {
  description = "Canonical host of the site; www.<domain> redirects to it."
  type        = string
}

variable "zone_name" {
  description = "Route 53 hosted zone that domain_name lives in."
  type        = string
}
