(() => {
  "use strict";

  document.querySelectorAll("[data-copy-target]").forEach((button) => {
    button.addEventListener("click", async () => {
      const target = document.querySelector(button.dataset.copyTarget || "");
      const status = document.querySelector("[data-copy-status]");
      if (!target || !status) return;
      try {
        await navigator.clipboard.writeText(target.textContent || "");
        status.textContent = "Installation command copied.";
      } catch (_) {
        status.textContent = "Copy was unavailable; select the command manually.";
      }
    });
  });

  const config = window.__SITE_CONFIG__;
  const updateNotice = document.querySelector("[data-pwa-update]");
  const updateButton = document.querySelector("[data-pwa-update-now]");
  const secureLocal = location.hostname === "127.0.0.1";
  if (!config?.enablePWA || !(location.protocol === "https:" || secureLocal) || !("serviceWorker" in navigator)) return;

  let refreshing = false;
  navigator.serviceWorker.addEventListener("controllerchange", () => {
    if (refreshing) return;
    refreshing = true;
    location.reload();
  });

  navigator.serviceWorker.register(config.serviceWorkerURL, { scope: config.basePath }).then((registration) => {
    const showUpdate = (worker) => {
      if (!navigator.serviceWorker.controller || !worker) return;
      updateNotice.hidden = false;
      updateButton.onclick = () => worker.postMessage({ type: "SKIP_WAITING" });
    };
    showUpdate(registration.waiting);
    registration.addEventListener("updatefound", () => {
      registration.installing?.addEventListener("statechange", () => {
        if (registration.installing?.state === "installed") showUpdate(registration.waiting);
      });
    });
  }).catch(() => {
    // Website loading remains independent of optional service-worker registration.
  });
})();
