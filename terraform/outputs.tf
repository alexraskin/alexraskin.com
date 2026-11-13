output "cloudfront_url" {
  value = "https://${module.cdn.domain_name}"
}

output "distribution_id" {
  value = module.cdn.id
}

output "github_deploy_role_arn" {
  value = module.github_oidc.role_arn
}
