// CloudFront Function (viewer request, cloudfront-js-2.0).
// Hugo pretty URLs: /dir/ -> /dir/index.html; /dir -> 301 /dir/ so relative links resolve.
function handler(event) {
  var r = event.request;
  if (r.uri.endsWith("/")) {
    r.uri += "index.html";
    return r;
  }
  if (r.uri.split("/").pop().indexOf(".") === -1) {
    var qs = Object.keys(r.querystring).map(function (k) {
      return k + "=" + r.querystring[k].value;
    }).join("&");
    return {
      statusCode: 301,
      statusDescription: "Moved Permanently",
      headers: { location: { value: r.uri + "/" + (qs ? "?" + qs : "") } },
    };
  }
  return r;
}
