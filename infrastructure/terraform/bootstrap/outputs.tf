output "role_arns" {
  value = { for env, r in module.deploy_role : env => r.arn }
}
