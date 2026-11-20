output "cloudfront_url" {
  value = "https://${module.cdn.domain_name}"
}

output "distribution_id" {
  value = module.cdn.id
}

output "bucket_name" {
  value = module.site_bucket.id
}
