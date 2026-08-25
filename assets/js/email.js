// The address is assembled at runtime instead of being written into the
// markup, so it never appears in the served HTML for harvesters to grep.
// A scraper driving a real browser still gets it; that is the ceiling here.
(function () {
  var el = document.getElementById("email");
  if (!el) {
    return;
  }
  var user = "hi";
  var domain = ["alexraskin", "com"].join(".");
  el.textContent = user + String.fromCharCode(64) + domain;
})();
