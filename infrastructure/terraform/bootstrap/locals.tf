locals {
  tags = {
    Project     = "alexraskin.com"
    Environment = "bootstrap"
    ManagedBy   = "terraform"
  }
  state_bucket = "alexraskin-tfstate"
  site_bucket  = "alexraskin-website"
  zone         = "alexraskin.com"

  # ACM validation names are _<32 hex>.<host>; _dmarc and friends stay out of reach.
  acm_label = "_${join("", [for i in range(32) : "?"])}"
  dns_names = flatten([for h in [local.zone, "www.${local.zone}"] : [h, "${local.acm_label}.${h}"]])
}
