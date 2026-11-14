locals {
  hostnames = [var.domain, "www.${var.domain}"]

  # Cloudflare has no default tags; every resource with a `comment` gets this.
  cloudflare_comment = "managed-by:terraform;application:alexraskin.com"
}
