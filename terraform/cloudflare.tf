provider "cloudflare" {
  api_token = var.cloudflare_token # Zone:DNS:Edit + Zone:Read
}

module "dns" {
  source = "./modules/cloudflare-dns"
  count  = var.attach_domain ? 1 : 0

  zone_name              = var.domain
  hostnames              = local.hostnames
  target                 = module.cdn.domain_name
  acm_validation_options = tolist(aws_acm_certificate.site[0].domain_validation_options)
}
