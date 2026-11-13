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

resource "cloudflare_dns_record" "acm" {
  for_each = { for o in var.acm_validation_options : o.domain_name => o }

  zone_id = data.cloudflare_zone.site.id
  name    = trimsuffix(each.value.resource_record_name, ".")
  type    = each.value.resource_record_type
  content = trimsuffix(each.value.resource_record_value, ".")
  ttl     = 60
  proxied = false
}

# Unproxied: CloudFront is the CDN; Cloudflare only answers DNS. Apex CNAME is flattened by Cloudflare.
resource "cloudflare_dns_record" "site" {
  for_each = toset(var.hostnames)

  zone_id = data.cloudflare_zone.site.id
  name    = each.value
  type    = "CNAME"
  content = var.target
  ttl     = 300
  proxied = false
}

output "acm_record_names" {
  value = [for r in cloudflare_dns_record.acm : r.name]
}
