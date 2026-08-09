package site

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testConfig(t *testing.T) SiteConfig {
	t.Helper()
	root := testProjectRoot(t)
	cfg, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	return cfg
}

func testProjectRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func temporaryConfigProject(t *testing.T, modulePath, goVersion string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(testProjectRoot(t), "site.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "site.config.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+modulePath+"\n\ngo "+goVersion+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestLoadConfigDerivations(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	if cfg.PagesBasePath != "/gurl/" {
		t.Fatalf("PagesBasePath = %q", cfg.PagesBasePath)
	}
	if cfg.PagesURL != "https://chengchuu.github.io/gurl/" {
		t.Fatalf("PagesURL = %q", cfg.PagesURL)
	}
	if cfg.Theme.StorageKey != "gurl-theme" || cfg.CachePrefix != "gurl-site-" {
		t.Fatalf("target keys = %q and %q", cfg.Theme.StorageKey, cfg.CachePrefix)
	}
}

func TestPagesHaveUniqueMetadata(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	titles := make(map[string]bool)
	descriptions := make(map[string]bool)
	for _, page := range cfg.Pages {
		if titles[page.Title] {
			t.Errorf("duplicate page title %q", page.Title)
		}
		titles[page.Title] = true
		if descriptions[page.Description] {
			t.Errorf("duplicate page description %q", page.Description)
		}
		descriptions[page.Description] = true
	}
}

func TestConfiguredThemePalette(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	want := ThemeConfig{
		StorageKey: "gurl-theme",
		ColorLight: "#ffffff",
		ColorDark:  "#141414",
		Primary: ThemePalette{
			Light: PrimaryPalette{Base: "#4d8ffb", Hover: "#256fd8", Active: "#185aaa", Soft: "#eaf2ff", RGB: "77, 143, 251", HoverRGB: "37, 111, 216", ActiveRGB: "24, 90, 170"},
			Dark:  PrimaryPalette{Base: "#5089e8", Hover: "#6198ee", Active: "#74a5f3", Soft: "#1b3155", RGB: "80, 137, 232", HoverRGB: "97, 152, 238", ActiveRGB: "116, 165, 243"},
		},
	}
	if cfg.Theme != want {
		t.Fatalf("theme = %+v, want %+v", cfg.Theme, want)
	}
}

func TestLoadConfigRejectsUnknownFields(t *testing.T) {
	t.Parallel()
	root := testProjectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "site.config.json"))
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "{", "{\n  \"unexpected\": true,", 1))
	temporaryRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(temporaryRoot, "site.config.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(temporaryRoot); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("LoadConfig error = %v, want unknown field", err)
	}
}

func TestLoadConfigRejectsModulePathMismatch(t *testing.T) {
	t.Parallel()
	root := temporaryConfigProject(t, "example.test/different-module", "1.25.0")
	if _, err := LoadConfig(root); err == nil || !strings.Contains(err.Error(), "does not match go.mod") {
		t.Fatalf("LoadConfig error = %v, want module path mismatch", err)
	}
}

func TestLoadConfigRejectsGoVersionMismatch(t *testing.T) {
	t.Parallel()
	root := temporaryConfigProject(t, "github.com/chengchuu/gurl", "1.20")
	if _, err := LoadConfig(root); err == nil || !strings.Contains(err.Error(), "does not match go.mod Go version") {
		t.Fatalf("LoadConfig error = %v, want Go version mismatch", err)
	}
}

func TestValidateRejectsUnsafeRepositoryIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*SiteConfig)
	}{
		{name: "owner changes URL authority", mutate: func(cfg *SiteConfig) { cfg.RepositoryOwner = "owner@example.com" }},
		{name: "repository traverses path", mutate: func(cfg *SiteConfig) { cfg.RepositoryName = "../outside" }},
		{name: "repository has surrounding whitespace", mutate: func(cfg *SiteConfig) { cfg.RepositoryName = " repository " }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := testConfig(t)
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate accepted unsafe repository identity")
			}
		})
	}
}

func TestValidateRejectsDerivedIdentityDrift(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*SiteConfig)
	}{
		{name: "repository URL", mutate: func(cfg *SiteConfig) { cfg.RepositoryURL = "https://github.com/other/project" }},
		{name: "theme storage key", mutate: func(cfg *SiteConfig) { cfg.Theme.StorageKey = "shared-theme" }},
		{name: "cache prefix", mutate: func(cfg *SiteConfig) { cfg.CachePrefix = "shared-pages-" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := testConfig(t)
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate accepted derived identity drift")
			}
		})
	}
}

func TestValidateRejectsMissingPublicationFields(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*SiteConfig)
	}{
		{name: "Go version", mutate: func(cfg *SiteConfig) { cfg.GoVersion = "" }},
		{name: "license", mutate: func(cfg *SiteConfig) { cfg.License = "" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := testConfig(t)
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate accepted a missing publication field")
			}
		})
	}
}

func TestValidateRejectsUnsafeIconFilenames(t *testing.T) {
	t.Parallel()
	for _, filename := range []string{"../favicon.png", `nested\favicon.png`, "favicon.svg"} {
		t.Run(filename, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.Favicon = filename
			if err := cfg.Validate(); err == nil {
				t.Fatalf("Validate accepted unsafe icon filename %q", filename)
			}
		})
	}
}

func TestValidateRejectsBrokenStablePageContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*SiteConfig)
	}{
		{name: "output traversal", mutate: func(cfg *SiteConfig) { cfg.Pages[0].OutputPath = "../outside.html" }},
		{name: "changed route", mutate: func(cfg *SiteConfig) { cfg.Pages[1].Path = "samples/" }},
		{name: "missing schema", mutate: func(cfg *SiteConfig) { cfg.Pages[2].SchemaType = "" }},
		{name: "duplicate name", mutate: func(cfg *SiteConfig) { cfg.Pages[1].Name = "home" }},
		{name: "duplicate title", mutate: func(cfg *SiteConfig) { cfg.Pages[1].Title = cfg.Pages[0].Title }},
		{name: "duplicate description", mutate: func(cfg *SiteConfig) { cfg.Pages[1].Description = cfg.Pages[0].Description }},
		{name: "padded title", mutate: func(cfg *SiteConfig) { cfg.Pages[1].Title = " Examples " }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := testConfig(t)
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate accepted a broken stable page contract")
			}
		})
	}
}

func TestDerivePagesBasePath(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"library":        "/library/",
		" team library ": "/team%20library/",
		"":               "/",
		"../unsafe":      "/",
	}
	for input, want := range tests {
		if got := DerivePagesBasePath(input); got != want {
			t.Errorf("DerivePagesBasePath(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestCanonicalURLs(t *testing.T) {
	t.Parallel()
	cfg := testConfig(t)
	want := map[string]string{
		"home":     "https://chengchuu.github.io/gurl/",
		"examples": "https://chengchuu.github.io/gurl/examples/",
		"api":      "https://chengchuu.github.io/gurl/api/",
	}
	for name, expected := range want {
		page, ok := cfg.Page(name)
		if !ok {
			t.Fatalf("page %q not found", name)
		}
		if got := cfg.CanonicalURL(page); got != expected {
			t.Errorf("canonical %s = %q, want %q", name, got, expected)
		}
	}
}
