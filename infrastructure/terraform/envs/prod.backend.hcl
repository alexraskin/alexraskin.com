# Shared state bucket across all projects.
bucket       = "alexraskin-tfstate"
key          = "alexraskin/prod/terraform.tfstate"
region       = "us-east-1"
use_lockfile = true
encrypt      = true
