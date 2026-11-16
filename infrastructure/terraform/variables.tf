variable "environment" {
  type = string
  validation {
    condition     = contains(["prod"], var.environment)
    error_message = "environment must be prod."
  }
}

variable "aws_region" {
  type        = string
  description = "Must be us-east-1: CloudFront only accepts ACM certs from there."
}

variable "bucket_name" {
  type    = string
  default = "alexraskin-website"
}

variable "domain" {
  type    = string
  default = "alexraskin.com"
}

variable "attach_domain" {
  type    = bool
  default = true
}

variable "cloudflare_token" {
  type      = string
  default   = null
  sensitive = true
}
