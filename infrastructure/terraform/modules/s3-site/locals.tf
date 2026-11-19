locals {
  files = fileset(var.site_dir, "**")

  mime = {
    html        = "text/html; charset=utf-8"
    css         = "text/css"
    js          = "text/javascript"
    woff2       = "font/woff2"
    webp        = "image/webp"
    jpg         = "image/jpeg"
    jpeg        = "image/jpeg"
    gif         = "image/gif"
    svg         = "image/svg+xml"
    png         = "image/png"
    ico         = "image/x-icon"
    txt         = "text/plain; charset=utf-8"
    xml         = "application/xml; charset=utf-8"
    json        = "application/json"
    webmanifest = "application/manifest+json"
  }

  cache_by_prefix = {
    fonts = "public, max-age=31536000, immutable"
    css   = "public, max-age=31536000, immutable" # Hugo fingerprints these
    img   = "public, max-age=86400, stale-while-revalidate=604800"
    icons = "public, max-age=86400, stale-while-revalidate=604800"
  }
  cache_default = "public, max-age=0, must-revalidate" # HTML: deploys show up immediately
}
