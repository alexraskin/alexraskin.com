resource "aws_route53_record" "acm" {
  for_each = toset(var.hostnames)

  zone_id = data.aws_route53_zone.site.zone_id
  name    = local.acm[each.key].resource_record_name
  type    = local.acm[each.key].resource_record_type
  records = [local.acm[each.key].resource_record_value]
  ttl     = 60
}

# Apex can't be a CNAME in Route 53, so every hostname aliases the distribution.
resource "aws_route53_record" "site" {
  for_each = { for p in setproduct(var.hostnames, ["A", "AAAA"]) : "${p[0]} ${p[1]}" => p }

  zone_id = data.aws_route53_zone.site.zone_id
  name    = each.value[0]
  type    = each.value[1]

  alias {
    name                   = var.target.domain_name
    zone_id                = var.target.hosted_zone_id
    evaluate_target_health = false
  }
}
