# GitHub Actions -> AWS via OIDC (no long-lived keys).
# CI role has read-only IAM, so changes to this file must be applied locally.

variable "repo" {
  type        = string
  description = "owner/name; only its main branch may assume the role."
}

variable "role_name" {
  type = string
}

variable "state_bucket" {
  type = string
}

variable "state_key_prefix" {
  type = string
}

variable "site_bucket_arn" {
  type = string
}

resource "aws_iam_openid_connect_provider" "github" {
  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]
}

resource "aws_iam_role" "github_deploy" {
  name = var.role_name
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Federated = aws_iam_openid_connect_provider.github.arn }
      Action    = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "token.actions.githubusercontent.com:aud" = "sts.amazonaws.com"
          "token.actions.githubusercontent.com:sub" = "repo:${var.repo}:ref:refs/heads/main"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "github_deploy" {
  role = aws_iam_role.github_deploy.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid      = "StateList"
        Effect   = "Allow"
        Action   = "s3:ListBucket"
        Resource = "arn:aws:s3:::${var.state_bucket}"
      },
      {
        Sid      = "StateRW"
        Effect   = "Allow"
        Action   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
        Resource = "arn:aws:s3:::${var.state_bucket}/${var.state_key_prefix}/*"
      },
      {
        Sid      = "SiteBucket"
        Effect   = "Allow"
        Action   = "s3:*"
        Resource = [var.site_bucket_arn, "${var.site_bucket_arn}/*"]
      },
      {
        # OAC, functions, cache policies and new certs can't be resource-scoped up front.
        Sid      = "CdnAndCerts"
        Effect   = "Allow"
        Action   = ["cloudfront:*", "acm:*"]
        Resource = "*"
      },
      {
        Sid      = "ReadBudgets"
        Effect   = "Allow"
        Action   = ["budgets:ViewBudget", "budgets:ListTagsForResource"]
        Resource = "*"
      },
      {
        Sid    = "RefreshOwnIam"
        Effect = "Allow"
        Action = [
          "iam:GetOpenIDConnectProvider",
          "iam:GetRole",
          "iam:GetRolePolicy",
          "iam:ListRolePolicies",
          "iam:ListAttachedRolePolicies",
        ]
        Resource = [aws_iam_openid_connect_provider.github.arn, aws_iam_role.github_deploy.arn]
      },
    ]
  })
}

output "role_arn" {
  value = aws_iam_role.github_deploy.arn
}
