(() => {
  "use strict";

  const config = window.__SITE_CONFIG__;
  if (!config) return;
  const media = window.matchMedia("(prefers-color-scheme: dark)");
  const allowed = new Set(["system", "light", "dark"]);

  function readPreference() {
    try {
      const value = localStorage.getItem(config.themeStorageKey) || "system";
      return allowed.has(value) ? value : "system";
    } catch (_) {
      return "system";
    }
  }

  function resolve(preference) {
    return preference === "system" ? (media.matches ? "dark" : "light") : preference;
  }

  function apply(preference) {
    const resolved = resolve(preference);
    document.documentElement.dataset.bsTheme = resolved;
    document.documentElement.dataset.themePreference = preference;
    document.querySelectorAll("meta[data-theme-color]").forEach((meta) => {
      meta.content = resolved === "dark" ? config.themeDark : config.themeLight;
    });
    document.querySelectorAll("[data-theme-select]").forEach((select) => {
      select.value = preference;
    });
  }

  function store(preference) {
    try {
      localStorage.setItem(config.themeStorageKey, preference);
    } catch (_) {}
  }

  document.querySelectorAll("[data-theme-select]").forEach((select) => {
    select.addEventListener("change", () => {
      const preference = allowed.has(select.value) ? select.value : "system";
      store(preference);
      apply(preference);
    });
  });

  media.addEventListener("change", () => {
    if (readPreference() === "system") apply("system");
  });
  apply(readPreference());
})();
