module "github_oidc" {
  source = "./modules/github-oidc"

  repo             = "alexraskin/alexraskin.com"
  role_name        = "github-deploy-website"
  state_bucket     = "alexraskin-tfstate"
  state_key_prefix = "website"
  site_bucket_arn  = module.site_bucket.arn
}
