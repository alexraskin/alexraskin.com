output "role_arns" {
  value = { for env, r in aws_iam_role.deploy : env => r.arn }
}
