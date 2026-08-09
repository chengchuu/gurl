package site

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// BuildResult describes a completed deterministic Pages build.
type BuildResult struct {
	OutputDir string
	Pages     int
	Assets    int
	Version   string
}

type runtimeConfig struct {
	BasePath         string `json:"basePath"`
	ThemeStorageKey  string `json:"themeStorageKey"`
	ThemeLight       string `json:"themeLight"`
	ThemeDark        string `json:"themeDark"`
	ServiceWorkerURL string `json:"serviceWorkerURL"`
	EnablePWA        bool   `json:"enablePWA"`
}

type templateData struct {
	Config       SiteConfig
	Page         PageConfig
	Metadata     PageMetadata
	API          APIDocumentation
	BasicExample string
	ExampleTest  string
	Navigation   template.HTML
	Content      template.HTML
	RuntimeJSON  template.JS
	ManifestURL  string
	FaviconURL   string
	AppleIconURL string
	SitemapURL   string
}

// Build creates the complete GitHub Pages artifact from source templates and the public Go API.
func Build(projectRoot string, cfg SiteConfig) (BuildResult, error) {
	if err := cfg.Validate(); err != nil {
		return BuildResult{}, fmt.Errorf("validate site configuration: %w", err)
	}
	if err := ValidateSourceIdentity(projectRoot, cfg); err != nil {
		return BuildResult{}, err
	}
	outputDir := filepath.Join(projectRoot, cfg.OutputDir)
	sourceDir := filepath.Join(projectRoot, cfg.SiteSourceDir)
	api, err := ExtractAPI(projectRoot, cfg.ModulePath)
	if err != nil {
		return BuildResult{}, err
	}
	if api.PackageName != cfg.PackageName {
		return BuildResult{}, fmt.Errorf("configured package name %q does not match public Go package %q", cfg.PackageName, api.PackageName)
	}
	basicExample, err := os.ReadFile(filepath.Join(projectRoot, "examples", "basic", "main.go"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("read basic example: %w", err)
	}
	exampleTest, err := os.ReadFile(filepath.Join(projectRoot, "example_test.go"))
	if err != nil {
		return BuildResult{}, fmt.Errorf("read example test: %w", err)
	}
	if err := os.RemoveAll(outputDir); err != nil {
		return BuildResult{}, fmt.Errorf("replace generated output: %w", err)
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return BuildResult{}, fmt.Errorf("create output directory: %w", err)
	}

	assetGroups := []string{"css", "js", "vendor"}
	for _, group := range assetGroups {
		if err := copyTree(filepath.Join(sourceDir, "assets", group), filepath.Join(outputDir, "assets", group)); err != nil {
			return BuildResult{}, err
		}
	}
	if err := copyTree(filepath.Join(sourceDir, "assets", "icons"), filepath.Join(outputDir, "icons")); err != nil {
		return BuildResult{}, err
	}

	for _, page := range cfg.Pages {
		if err := renderPage(sourceDir, outputDir, cfg, page, api, string(basicExample), string(exampleTest)); err != nil {
			return BuildResult{}, err
		}
	}

	manifest, err := Manifest(cfg)
	if err != nil {
		return BuildResult{}, fmt.Errorf("generate manifest: %w", err)
	}
	if err := writeFile(filepath.Join(outputDir, "manifest.webmanifest"), append(manifest, '\n')); err != nil {
		return BuildResult{}, err
	}
	if err := writeFile(filepath.Join(outputDir, "robots.txt"), []byte(Robots(cfg))); err != nil {
		return BuildResult{}, err
	}
	if err := writeFile(filepath.Join(outputDir, "sitemap.xml"), []byte(Sitemap(cfg))); err != nil {
		return BuildResult{}, err
	}
	if err := writeFile(filepath.Join(outputDir, ".nojekyll"), nil); err != nil {
		return BuildResult{}, err
	}

	precache, err := precacheURLs(outputDir, cfg)
	if err != nil {
		return BuildResult{}, err
	}
	version, err := contentVersion(outputDir)
	if err != nil {
		return BuildResult{}, err
	}
	worker := ServiceWorker(cfg, version, precache)
	if err := writeFile(filepath.Join(outputDir, "service-worker.js"), []byte(worker)); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{OutputDir: outputDir, Pages: len(cfg.Pages), Assets: len(precache), Version: version}, nil
}

func renderPage(sourceDir, outputDir string, cfg SiteConfig, page PageConfig, api APIDocumentation, basicExample, exampleTest string) error {
	metadata, err := Metadata(cfg, page)
	if err != nil {
		return err
	}
	runtimeBytes, err := json.Marshal(runtimeConfig{
		BasePath: cfg.PagesBasePath, ThemeStorageKey: cfg.Theme.StorageKey,
		ThemeLight: cfg.Theme.ColorLight, ThemeDark: cfg.Theme.ColorDark,
		ServiceWorkerURL: cfg.AssetURL("service-worker.js"), EnablePWA: true,
	})
	if err != nil {
		return fmt.Errorf("encode runtime configuration: %w", err)
	}
	data := templateData{
		Config: cfg, Page: page, Metadata: metadata, API: api,
		BasicExample: basicExample, ExampleTest: exampleTest,
		RuntimeJSON:  template.JS(runtimeBytes),
		ManifestURL:  cfg.AssetURL("manifest.webmanifest"),
		FaviconURL:   cfg.AssetURL("icons/" + cfg.Favicon),
		AppleIconURL: cfg.AssetURL("icons/" + cfg.AppleTouchIcon),
		SitemapURL:   cfg.PagesURL + "sitemap.xml",
	}
	templateDir := filepath.Join(sourceDir, "templates")
	navigation, err := executeTemplate(filepath.Join(templateDir, "navigation.tmpl"), "navigation", data)
	if err != nil {
		return err
	}
	data.Navigation = template.HTML(navigation)
	content, err := executeTemplate(filepath.Join(templateDir, page.Name+".tmpl"), "content", data)
	if err != nil {
		return err
	}
	data.Content = template.HTML(content)
	layout, err := template.ParseFiles(filepath.Join(templateDir, "layout.tmpl"))
	if err != nil {
		return fmt.Errorf("parse layout template: %w", err)
	}
	outputPath := filepath.Join(outputDir, filepath.FromSlash(page.OutputPath))
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return fmt.Errorf("create page directory: %w", err)
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create page %s: %w", page.Name, err)
	}
	defer file.Close()
	if err := layout.ExecuteTemplate(file, "layout", data); err != nil {
		return fmt.Errorf("render page %s: %w", page.Name, err)
	}
	return nil
}

func executeTemplate(filename, name string, data templateData) (string, error) {
	parsed, err := template.ParseFiles(filename)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", filename, err)
	}
	var output strings.Builder
	if err := parsed.ExecuteTemplate(&output, name, data); err != nil {
		return "", fmt.Errorf("render template %s: %w", filename, err)
	}
	return output.String(), nil
}

func copyTree(source, destination string) error {
	return filepath.WalkDir(source, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, sourcePath)
		if err != nil {
			return err
		}
		destinationPath := filepath.Join(destination, relative)
		if entry.IsDir() {
			return os.MkdirAll(destinationPath, 0o755)
		}
		input, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer input.Close()
		output, err := os.Create(destinationPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(output, input); err != nil {
			output.Close()
			return err
		}
		return output.Close()
	})
}

func writeFile(filename string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	return nil
}

func precacheURLs(outputDir string, cfg SiteConfig) ([]string, error) {
	urls := []string{cfg.PagesBasePath, cfg.AssetURL("examples/"), cfg.AssetURL("api/")}
	err := filepath.WalkDir(outputDir, func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		relative, err := filepath.Rel(outputDir, filename)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if relative == "index.html" || relative == "examples/index.html" || relative == "api/index.html" || relative == ".nojekyll" || relative == "robots.txt" || relative == "sitemap.xml" {
			return nil
		}
		urls = append(urls, cfg.AssetURL(relative))
		return nil
	})
	sort.Strings(urls)
	return urls, err
}

func contentVersion(outputDir string) (string, error) {
	var names []string
	if err := filepath.WalkDir(outputDir, func(filename string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		names = append(names, filename)
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		relative, _ := filepath.Rel(outputDir, name)
		hash.Write([]byte(filepath.ToSlash(relative)))
		data, err := os.ReadFile(name)
		if err != nil {
			return "", err
		}
		hash.Write(data)
	}
	return hex.EncodeToString(hash.Sum(nil))[:12], nil
}
