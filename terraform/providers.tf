provider "aws" {
  region = "us-east-1" # CloudFront certs must live in us-east-1

  default_tags {
    tags = {
      Project   = var.domain
      ManagedBy = "terraform"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_token # Zone:DNS:Edit + Zone:Read
}
