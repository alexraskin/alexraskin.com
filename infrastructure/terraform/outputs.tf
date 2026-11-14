output "cloudfront_url" {
  value = "https://${module.cdn.domain_name}"
}

output "distribution_id" {
  value = module.cdn.id
}
