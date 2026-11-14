# Named outside the site's resource names so the role can't edit its own policy.
resource "aws_iam_role" "deploy" {
  for_each = var.environments

  name                 = "github-deploy-website-${each.key}"
  description          = "GitHub Actions deploy for ${var.github_repository} (${each.key})"
  assume_role_policy   = data.aws_iam_policy_document.trust[each.key].json
  max_session_duration = 3600
}

resource "aws_iam_role_policy" "deploy" {
  for_each = var.environments

  name   = "deploy"
  role   = aws_iam_role.deploy[each.key].id
  policy = data.aws_iam_policy_document.deploy[each.key].json
}
