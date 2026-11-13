provider "aws" {
  region = "us-east-1" # CloudFront certs must live in us-east-1
}

module "site_bucket" {
  source = "./modules/s3-site"

  bucket_name = var.bucket_name
  site_dir    = "${path.module}/../public"
}

module "cdn" {
  source = "./modules/cloudfront"

  name                        = var.bucket_name
  bucket_id                   = module.site_bucket.id
  bucket_arn                  = module.site_bucket.arn
  bucket_regional_domain_name = module.site_bucket.regional_domain_name
  aliases                     = var.attach_domain ? local.hostnames : []
  acm_certificate_arn         = var.attach_domain ? aws_acm_certificate_validation.site[0].certificate_arn : null
}

resource "aws_acm_certificate" "site" {
  count                     = var.attach_domain ? 1 : 0
  domain_name               = var.domain
  subject_alternative_names = ["www.${var.domain}"]
  validation_method         = "DNS"

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "site" {
  count                   = var.attach_domain ? 1 : 0
  certificate_arn         = aws_acm_certificate.site[0].arn
  validation_record_fqdns = module.dns[0].acm_record_names
}
