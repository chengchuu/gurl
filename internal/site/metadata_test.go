package site

import (
	"strings"
	"testing"
)

func TestMetadataIsPageSpecific(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	seen := make(map[string]bool)
	for _, page := range cfg.Pages {
		metadata, err := Metadata(cfg, page)
		if err != nil {
			t.Fatal(err)
		}
		if metadata.Canonical != cfg.CanonicalURL(page) {
			t.Errorf("canonical for %s = %q", page.Name, metadata.Canonical)
		}
		if seen[metadata.Canonical] {
			t.Errorf("duplicate canonical %q", metadata.Canonical)
		}
		seen[metadata.Canonical] = true
		if !strings.Contains(string(metadata.JSONLD), metadata.Canonical) {
			t.Errorf("JSON-LD for %s omits canonical URL", page.Name)
		}
		if want := cfg.PagesURL + "icons/" + cfg.OpenGraphImage; metadata.OGImage != want {
			t.Errorf("Open Graph image for %s = %q, want %q", page.Name, metadata.OGImage, want)
		}
	}
}

func TestCrawlerFiles(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	robots := Robots(cfg)
	if want := "Sitemap: https://chengchuu.github.io/gurl/sitemap.xml"; !strings.Contains(robots, want) {
		t.Fatalf("robots.txt missing %q", want)
	}
	sitemap := Sitemap(cfg)
	if got := strings.Count(sitemap, "<loc>"); got != 3 {
		t.Fatalf("sitemap locations = %d, want 3", got)
	}
	for _, page := range cfg.Pages {
		if !strings.Contains(sitemap, "<loc>"+cfg.CanonicalURL(page)+"</loc>") {
			t.Errorf("sitemap missing %s", cfg.CanonicalURL(page))
		}
	}
}
