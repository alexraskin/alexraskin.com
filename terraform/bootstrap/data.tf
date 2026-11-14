# Trust: only jobs in this repo running in the GitHub environment named <env>.
data "aws_iam_policy_document" "trust" {
  for_each = var.environments

  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = [aws_iam_openid_connect_provider.github.arn]
    }
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["${var.github_sub_prefix}:environment:${each.key}"]
    }
  }
}

# Terraform apply + site upload. No IAM permissions: the role can't touch itself.
data "aws_iam_policy_document" "deploy" {
  for_each = var.environments

  statement {
    sid       = "StateList"
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::${local.state_bucket}"]
    condition {
      test     = "StringLike"
      variable = "s3:prefix"
      values   = ["website/${each.key}/*"]
    }
  }

  statement {
    sid       = "StateObjects"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = ["arn:aws:s3:::${local.state_bucket}/website/${each.key}/*"]
  }

  statement {
    sid     = "SiteBucket"
    actions = ["s3:*"]
    resources = [
      "arn:aws:s3:::${local.site_bucket}",
      "arn:aws:s3:::${local.site_bucket}/*",
    ]
  }

  # ponytail: CloudFront distributions/policies and ACM certs are addressed by
  # generated IDs, so they can't be name-scoped. Tag conditions if this matters.
  statement {
    sid       = "CdnAndCerts"
    actions   = ["cloudfront:*", "acm:*"]
    resources = ["*"]
  }
}
