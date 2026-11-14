// CloudFront Function (viewer request, cloudfront-js-2.0).
// Pretty URLs like Cloudflare assets: /cosmo -> /cosmo.html, /dir/ -> /dir/index.html.
function handler(event) {
  var r = event.request;
  if (r.uri.endsWith("/")) r.uri += "index.html";
  else if (r.uri.split("/").pop().indexOf(".") === -1) r.uri += ".html";
  return r;
}
