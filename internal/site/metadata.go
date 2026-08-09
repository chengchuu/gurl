package site

import (
	"encoding/json"
	"fmt"
	"html/template"
)

// PageMetadata contains the generated SEO and structured-data identity for a page.
type PageMetadata struct {
	Title       string
	Description string
	Canonical   string
	OGType      string
	SiteName    string
	OGImage     string
	OGImageAlt  string
	JSONLD      template.JS
}

// Metadata creates page-specific metadata from the central configuration.
func Metadata(cfg SiteConfig, page PageConfig) (PageMetadata, error) {
	if page.Title == "" || page.Description == "" {
		return PageMetadata{}, fmt.Errorf("metadata for %q is incomplete", page.Name)
	}
	canonical := cfg.CanonicalURL(page)
	jsonLD := map[string]any{
		"@context":    "https://schema.org",
		"@type":       page.SchemaType,
		"name":        page.Title,
		"description": page.Description,
		"url":         canonical,
		"isPartOf": map[string]string{
			"@type": "WebSite",
			"name":  cfg.DisplayName,
			"url":   cfg.PagesURL,
		},
	}
	if page.Name == "home" {
		jsonLD["codeRepository"] = cfg.RepositoryURL
		jsonLD["programmingLanguage"] = "Go"
		jsonLD["license"] = cfg.RepositoryURL + "/blob/main/LICENSE"
	}
	encoded, err := json.Marshal(jsonLD)
	if err != nil {
		return PageMetadata{}, fmt.Errorf("encode metadata: %w", err)
	}
	return PageMetadata{
		Title:       page.Title,
		Description: page.Description,
		Canonical:   canonical,
		OGType:      "website",
		SiteName:    cfg.DisplayName,
		OGImage:     cfg.PagesURL + "icons/" + cfg.OpenGraphImage,
		OGImageAlt:  cfg.DisplayName + " website preview",
		JSONLD:      template.JS(encoded), // The value is produced only by encoding/json.
	}, nil
}

// Robots renders the crawler policy for the production site.
func Robots(cfg SiteConfig) string {
	return "User-agent: *\nAllow: /\n\nSitemap: " + cfg.PagesURL + "sitemap.xml\n"
}

// Sitemap renders exactly the configured public page URLs.
func Sitemap(cfg SiteConfig) string {
	result := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<urlset xmlns=\"http://www.sitemaps.org/schemas/sitemap/0.9\">\n"
	for _, page := range cfg.Pages {
		result += "  <url><loc>" + cfg.CanonicalURL(page) + "</loc></url>\n"
	}
	return result + "</urlset>\n"
}
