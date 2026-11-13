terraform {
  required_providers {
    cloudflare = { source = "cloudflare/cloudflare" } # non-hashicorp source must be declared per module
  }
}

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

data "cloudflare_zone" "site" {
  filter = { name = var.zone_name }
}

locals {
  acm = { for o in var.acm_validation_options : o.domain_name => o }
}

# Keys must be known at plan; ACM options aren't until the cert exists. Assumes cert domains == hostnames.
resource "cloudflare_dns_record" "acm" {
  for_each   = toset(var.hostnames)
  depends_on = [cloudflare_dns_record.caa_amazon] # ACM checks CAA as soon as these appear

  zone_id = data.cloudflare_zone.site.id
  comment = "managed-by:terraform;application:alexraskin.com"
  name    = trimsuffix(local.acm[each.key].resource_record_name, ".")
  type    = local.acm[each.key].resource_record_type
  content = trimsuffix(local.acm[each.key].resource_record_value, ".")
  ttl     = 60
  proxied = false
}

# Unproxied: CloudFront is the CDN; Cloudflare only answers DNS. Apex CNAME is flattened by Cloudflare.
resource "cloudflare_dns_record" "site" {
  for_each = toset(var.hostnames)

  zone_id = data.cloudflare_zone.site.id
  comment = "managed-by:terraform;application:alexraskin.com"
  name    = each.value
  type    = "CNAME"
  content = var.target
  ttl     = 300
  proxied = false
}

# Cloudflare Universal SSL publishes CAA for its own CAs only; ACM fails with CAA_ERROR without this.
resource "cloudflare_dns_record" "caa_amazon" {
  zone_id = data.cloudflare_zone.site.id
  comment = "managed-by:terraform;application:alexraskin.com"
  name    = var.zone_name
  type    = "CAA"
  ttl     = 300
  data    = { flags = 0, tag = "issue", value = "amazon.com" }
}

output "acm_record_names" {
  value = [for r in cloudflare_dns_record.acm : r.name]
}
