locals {
  tags = merge(var.tags, {
    Project     = "alexraskin.com"
    Environment = var.environment
    ManagedBy   = "terraform"
  })
  name = "alexraskin-${var.environment}"
}
