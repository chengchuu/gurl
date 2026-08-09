package site

import (
	"encoding/json"
	"fmt"
	"strings"
)

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

type webManifest struct {
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	Description     string         `json:"description"`
	StartURL        string         `json:"start_url"`
	Scope           string         `json:"scope"`
	Display         string         `json:"display"`
	BackgroundColor string         `json:"background_color"`
	ThemeColor      string         `json:"theme_color"`
	Icons           []manifestIcon `json:"icons"`
}

// Manifest renders the installable website manifest from central identity.
func Manifest(cfg SiteConfig) ([]byte, error) {
	manifest := webManifest{
		Name: cfg.DisplayName, ShortName: cfg.PWAShortName, Description: cfg.Description,
		StartURL: cfg.PagesBasePath, Scope: cfg.PagesBasePath, Display: "standalone",
		BackgroundColor: cfg.Theme.ColorLight, ThemeColor: cfg.Theme.Primary.Light.Base,
		Icons: []manifestIcon{
			{Src: cfg.AssetURL("icons/" + cfg.Icon192), Sizes: "192x192", Type: "image/png", Purpose: "any"},
			{Src: cfg.AssetURL("icons/" + cfg.Icon512), Sizes: "512x512", Type: "image/png", Purpose: "any"},
			{Src: cfg.AssetURL("icons/" + cfg.MaskableIcon512), Sizes: "512x512", Type: "image/png", Purpose: "maskable"},
		},
	}
	return json.MarshalIndent(manifest, "", "  ")
}

// ServiceWorker generates a project-scoped worker with network-first HTML and cache-first assets.
func ServiceWorker(cfg SiteConfig, version string, precache []string) string {
	quoted := make([]string, len(precache))
	for index, item := range precache {
		quoted[index] = fmt.Sprintf("  %q", item)
	}
	return fmt.Sprintf(`"use strict";

const CACHE_PREFIX = %q;
const CACHE_VERSION = %q;
const PRECACHE = CACHE_PREFIX + "precache-" + CACHE_VERSION;
const RUNTIME = CACHE_PREFIX + "runtime-" + CACHE_VERSION;
const PROJECT_BASE = %q;
const PRECACHE_URLS = [
%s
];
const MAX_RUNTIME_ENTRIES = 64;

self.addEventListener("install", (event) => {
  event.waitUntil(caches.open(PRECACHE).then((cache) => cache.addAll(PRECACHE_URLS)));
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(keys.map((key) => {
      if (key.startsWith(CACHE_PREFIX) && key !== PRECACHE && key !== RUNTIME) {
        return caches.delete(key);
      }
      return undefined;
    }))).then(() => self.clients.claim()),
  );
});

self.addEventListener("message", (event) => {
  if (event.data && event.data.type === "SKIP_WAITING") self.skipWaiting();
});

function isCacheable(request, response) {
  if (!response || !response.ok || response.type !== "basic") return false;
  if (request.headers.has("authorization")) return false;
  const policy = response.headers.get("cache-control") || "";
  return !/(?:no-store|private)/i.test(policy);
}

async function trimRuntimeCache(cache) {
  const keys = await cache.keys();
  await Promise.all(keys.slice(0, Math.max(0, keys.length - MAX_RUNTIME_ENTRIES)).map((key) => cache.delete(key)));
}

async function networkFirst(request) {
  const cache = await caches.open(RUNTIME);
  try {
    const response = await fetch(request);
    if (isCacheable(request, response)) {
      await cache.put(request, response.clone());
      await trimRuntimeCache(cache);
    }
    return response;
  } catch (error) {
    return (await cache.match(request)) || (await caches.match(PROJECT_BASE)) || Response.error();
  }
}

async function cacheFirst(request) {
  const cached = await caches.match(request);
  if (cached) return cached;
  const response = await fetch(request);
  if (isCacheable(request, response)) {
    const cache = await caches.open(RUNTIME);
    await cache.put(request, response.clone());
    await trimRuntimeCache(cache);
  }
  return response;
}

self.addEventListener("fetch", (event) => {
  const request = event.request;
  if (request.method !== "GET") return;
  const url = new URL(request.url);
  if (url.origin !== self.location.origin || !url.pathname.startsWith(PROJECT_BASE)) return;
  const wantsHTML = request.mode === "navigate" || (request.headers.get("accept") || "").includes("text/html");
  event.respondWith(wantsHTML ? networkFirst(request) : cacheFirst(request));
});
`, cfg.CachePrefix, version, cfg.PagesBasePath, strings.Join(quoted, ",\n"))
}
