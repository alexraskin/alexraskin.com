variable "name" {
  type = string
}

variable "bucket_id" {
  type = string
}

variable "bucket_arn" {
  type = string
}

variable "bucket_regional_domain_name" {
  type = string
}

variable "aliases" {
  type    = list(string)
  default = []
}

variable "acm_certificate_arn" {
  type        = string
  default     = null
  description = "null = default *.cloudfront.net cert."
}
