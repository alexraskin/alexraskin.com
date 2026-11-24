output "site_url" {
  value = "https://${var.domain_name}"
}

output "bucket_name" {
  value = module.site.bucket_name
}

output "distribution_id" {
  value = module.site.distribution_id
}
