variable "bucket_name" {
  type    = string
  default = "alexraskin-website"
}

variable "domain" {
  type    = string
  default = "alexraskin.com"
}

# false = test on *.cloudfront.net only; flip to true to add ACM cert + Cloudflare DNS.
variable "attach_domain" {
  type    = bool
  default = false
}
