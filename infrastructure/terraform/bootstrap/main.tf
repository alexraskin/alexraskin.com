module "deploy_role" {
  source   = "git::https://github.com/alexraskin/terraform-hugo-module.git//modules/deploy_role?ref=v1.0.1"
  for_each = var.environments

  name              = "github-deploy-alexraskin-${each.key}"
  project           = local.tags.Project
  environment       = each.key
  github_repository = var.github_repository
  github_sub_prefix = var.github_sub_prefix
  state_key_prefix  = "alexraskin/${each.key}"
  resource_prefix   = "alexraskin-${each.key}-"
  route53_zone_name = local.zone
  hostnames         = [local.zone, "www.${local.zone}"]
}
