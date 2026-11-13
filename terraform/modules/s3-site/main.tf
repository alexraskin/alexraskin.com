variable "bucket_name" {
  type = string
}

variable "site_dir" {
  type        = string
  description = "Local directory uploaded to the bucket."
}

locals {
  files = setsubtract(fileset(var.site_dir, "**"), ["_headers"])

  mime = {
    html        = "text/html; charset=utf-8"
    css         = "text/css"
    js          = "text/javascript"
    woff2       = "font/woff2"
    webp        = "image/webp"
    jpg         = "image/jpeg"
    png         = "image/png"
    ico         = "image/x-icon"
    txt         = "text/plain; charset=utf-8"
    webmanifest = "application/manifest+json"
  }

  cache_by_prefix = {
    fonts = "public, max-age=31536000, immutable"
    js    = "public, max-age=31536000, immutable"
    css   = "public, max-age=3600, stale-while-revalidate=86400"
    img   = "public, max-age=86400, stale-while-revalidate=604800"
    icons = "public, max-age=86400, stale-while-revalidate=604800"
  }
  cache_default = "public, max-age=0, must-revalidate" # HTML: deploys show up immediately
}

resource "aws_s3_bucket" "site" {
  bucket = var.bucket_name
}

resource "aws_s3_bucket_public_access_block" "site" {
  bucket                  = aws_s3_bucket.site.id
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_object" "site" {
  for_each = local.files

  bucket        = aws_s3_bucket.site.id
  key           = each.value
  source        = "${var.site_dir}/${each.value}"
  etag          = filemd5("${var.site_dir}/${each.value}")
  content_type  = lookup(local.mime, reverse(split(".", each.value))[0], "application/octet-stream")
  cache_control = each.value == "favicon.ico" ? local.cache_by_prefix.icons : lookup(local.cache_by_prefix, split("/", each.value)[0], local.cache_default)
}

output "id" {
  value = aws_s3_bucket.site.id
}

output "arn" {
  value = aws_s3_bucket.site.arn
}

output "regional_domain_name" {
  value = aws_s3_bucket.site.bucket_regional_domain_name
}
