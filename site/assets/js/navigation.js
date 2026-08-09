(() => {
  "use strict";
  const navigation = document.querySelector("#site-navigation");
  if (!navigation || !window.bootstrap) return;
  navigation.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      if (window.getComputedStyle(document.querySelector(".navbar-toggler")).display !== "none") {
        window.bootstrap.Collapse.getOrCreateInstance(navigation, { toggle: false }).hide();
      }
    });
  });
})();
