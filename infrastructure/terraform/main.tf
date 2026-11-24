module "site" {
  source    = "git::https://github.com/alexraskin/terraform-hugo-module.git//modules/site?ref=v1.0.0"
  providers = { aws = aws, aws.us_east_1 = aws.us_east_1 }

  name                    = local.name
  bucket_name             = "${local.name}-web-${data.aws_caller_identity.current.account_id}"
  domain_name             = var.domain_name
  validation_record_fqdns = module.dns.validation_record_fqdns
  content_security_policy = "default-src 'self'; base-uri 'self'; object-src 'none'; form-action 'self'; frame-ancestors 'none'; img-src 'self'; font-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self'; connect-src 'self'"
}

module "dns" {
  source = "git::https://github.com/alexraskin/terraform-hugo-module.git//modules/route53_dns?ref=v1.0.0"

  zone_name                   = var.zone_name
  hostnames                   = module.site.hostnames
  acm_validation_records      = module.site.acm_validation_records
  distribution_domain_name    = module.site.distribution_domain_name
  distribution_hosted_zone_id = module.site.distribution_hosted_zone_id
}
