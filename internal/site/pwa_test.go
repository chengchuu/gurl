package site

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestManifestUsesProjectScope(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	data, err := Manifest(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest["start_url"] != cfg.PagesBasePath || manifest["scope"] != cfg.PagesBasePath {
		t.Fatalf("manifest scope = %v, start_url = %v", manifest["scope"], manifest["start_url"])
	}
	if manifest["theme_color"] != cfg.Theme.Primary.Light.Base || manifest["background_color"] != cfg.Theme.ColorLight {
		t.Fatalf("manifest theme = %v, background = %v", manifest["theme_color"], manifest["background_color"])
	}
}

func TestServiceWorkerIsProjectScoped(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	worker := ServiceWorker(cfg, "test-version", []string{cfg.PagesBasePath})
	for _, required := range []string{cfg.CachePrefix, cfg.PagesBasePath, "request.method !== \"GET\"", "url.origin !== self.location.origin", "networkFirst", "cacheFirst", "MAX_RUNTIME_ENTRIES"} {
		if !strings.Contains(worker, required) {
			t.Errorf("service worker missing %q", required)
		}
	}
	if strings.Contains(worker, "mazey-npm-template") {
		t.Fatal("service worker contains reference project identity")
	}
	if strings.Contains(worker, "|| url.search") {
		t.Fatal("service worker bypasses project navigations that contain a query string")
	}
}
