terraform {
  required_providers {
    cloudflare = { source = "cloudflare/cloudflare" } # non-hashicorp source must be declared per module
  }
}
