// Cookie consent for Google Analytics. GA is only loaded after the visitor
// clicks Accept; the choice is stored in localStorage and can be changed via
// any .consent-reopen button (footer, privacy page).
(function () {
  var KEY = "ga-consent";
  var script = document.currentScript;
  var gaId = script && script.dataset.gaId;
  var banner = document.getElementById("consent-banner");

  function read() {
    try { return localStorage.getItem(KEY); } catch (e) { return null; }
  }
  function write(v) {
    try { localStorage.setItem(KEY, v); } catch (e) {}
  }

  function loadGA() {
    if (!gaId || window.__gaLoaded) return;
    window.__gaLoaded = true;
    window.dataLayer = window.dataLayer || [];
    window.gtag = function () { window.dataLayer.push(arguments); };
    window.gtag("js", new Date());
    window.gtag("config", gaId);
    var s = document.createElement("script");
    s.async = true;
    s.src = "https://www.googletagmanager.com/gtag/js?id=" + encodeURIComponent(gaId);
    document.head.appendChild(s);
  }

  // Remove GA cookies after consent is withdrawn.
  function clearGACookies() {
    var host = location.hostname;
    var domains = ["", host, "." + host, "." + host.split(".").slice(-2).join(".")];
    document.cookie.split(";").forEach(function (c) {
      var name = c.split("=")[0].trim();
      if (name === "_ga" || name.indexOf("_ga_") === 0) {
        domains.forEach(function (d) {
          document.cookie = name + "=; Max-Age=0; path=/" + (d ? "; domain=" + d : "");
        });
      }
    });
  }

  function show() { if (banner) banner.hidden = false; }
  function hide() { if (banner) banner.hidden = true; }

  function choose(value) {
    var previous = read();
    write(value);
    hide();
    if (value === "granted") {
      loadGA();
    } else if (previous === "granted") {
      clearGACookies();
      location.reload(); // unload the already running GA script
    }
  }

  if (banner) {
    banner.addEventListener("click", function (e) {
      var v = e.target.getAttribute && e.target.getAttribute("data-consent");
      if (v) choose(v);
    });
  }
  document.addEventListener("click", function (e) {
    if (e.target.closest && e.target.closest(".consent-reopen")) show();
  });

  var stored = read();
  if (stored === "granted") loadGA();
  else if (stored !== "denied") show();
})();
