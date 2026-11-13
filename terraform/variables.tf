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
  default   = null # null = provider falls back to CLOUDFLARE_API_TOKEN env (CI)
  sensitive = true
}
