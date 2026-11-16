locals {
  hostnames = [var.domain, "www.${var.domain}"]

  cloudflare_comment = "managed-by:terraform;application:alexraskin.com"
}
