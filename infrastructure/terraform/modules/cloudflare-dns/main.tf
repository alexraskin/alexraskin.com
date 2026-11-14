# Keys must be known at plan; ACM options aren't until the cert exists. Assumes cert domains == hostnames.
resource "cloudflare_dns_record" "acm" {
  for_each   = toset(var.hostnames)
  depends_on = [cloudflare_dns_record.caa_amazon] # ACM checks CAA as soon as these appear

  zone_id = data.cloudflare_zone.site.id
  comment = var.cloudflare_comment
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
  comment = var.cloudflare_comment
  name    = each.value
  type    = "CNAME"
  content = var.target
  ttl     = 300
  proxied = false
}

# Cloudflare Universal SSL publishes CAA for its own CAs only; ACM fails with CAA_ERROR without this.
resource "cloudflare_dns_record" "caa_amazon" {
  zone_id = data.cloudflare_zone.site.id
  comment = var.cloudflare_comment
  name    = var.zone_name
  type    = "CAA"
  ttl     = 300
  data    = { flags = 0, tag = "issue", value = "amazon.com" }
}
