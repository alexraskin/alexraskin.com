provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Project     = var.domain
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
