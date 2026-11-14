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
