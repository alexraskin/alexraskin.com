# Set as AWS_ROLE_ARN on the GitHub environment of the same name.
output "role_arns" {
  value = { for env, r in aws_iam_role.deploy : env => r.arn }
}
