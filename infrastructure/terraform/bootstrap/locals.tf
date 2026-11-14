locals {
  tags = {
    Project     = "alexraskin.com"
    Environment = "bootstrap"
    ManagedBy   = "terraform"
  }
  state_bucket = "alexraskin-tfstate"
  site_bucket  = "alexraskin-website"
}
