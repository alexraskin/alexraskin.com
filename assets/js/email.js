(function () {
  var el = document.getElementById("email");
  if (!el) {
    return;
  }
  var user = "hi";
  var domain = ["alexraskin", "com"].join(".");
  el.textContent = user + String.fromCharCode(64) + domain;
})();
