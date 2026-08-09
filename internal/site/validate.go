package site

import (
	"encoding/json"
	"fmt"
	"html"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// ValidationResult summarizes final-artifact checks.
type ValidationResult struct {
	OutputDir string
	Pages     int
	Links     int
	Icons     int
}

var (
	attributePattern = regexp.MustCompile(`(?i)([a-zA-Z:-]+)\s*=\s*"([^"]*)"`)
	h1Pattern        = regexp.MustCompile(`(?is)<h1\b[^>]*>`)
	linkPattern      = regexp.MustCompile(`(?is)\b(?:href|src)="([^"]+)"`)
	tagPattern       = regexp.MustCompile(`(?is)<[^>]+>`)
	scriptPattern    = regexp.MustCompile(`(?is)<(?:script|style)\b[^>]*>.*?</(?:script|style)>`)
	locPattern       = regexp.MustCompile(`(?is)<loc>([^<]+)</loc>`)
)

// ValidateArtifact validates the generated Pages artifact rather than source templates.
func ValidateArtifact(projectRoot string, cfg SiteConfig) (ValidationResult, error) {
	if err := cfg.Validate(); err != nil {
		return ValidationResult{}, err
	}
	if err := ValidateSourceIdentity(projectRoot, cfg); err != nil {
		return ValidationResult{}, err
	}
	outputDir := filepath.Join(projectRoot, cfg.OutputDir)
	result := ValidationResult{OutputDir: outputDir}
	var failures []string
	pageHTML := make(map[string]string)

	for _, page := range cfg.Pages {
		filename := filepath.Join(outputDir, filepath.FromSlash(page.OutputPath))
		data, err := os.ReadFile(filename)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: required route is missing: %v", page.Name, err))
			continue
		}
		markup := string(data)
		pageHTML[page.OutputPath] = markup
		result.Pages++
		failures = append(failures, validateHTMLPage(cfg, page, markup)...)
	}

	for outputPath, markup := range pageHTML {
		count, linkFailures := validateLinks(outputDir, cfg, outputPath, markup)
		result.Links += count
		failures = append(failures, linkFailures...)
	}

	failures = append(failures, validateCrawlerFiles(outputDir, cfg)...)
	icons, pwaFailures := validatePWA(outputDir, cfg, pageHTML)
	result.Icons = icons
	failures = append(failures, pwaFailures...)
	failures = append(failures, FindUnresolvedPlaceholders(outputDir)...)

	if len(failures) > 0 {
		sort.Strings(failures)
		return result, fmt.Errorf("site validation failed:\n- %s", strings.Join(failures, "\n- "))
	}
	return result, nil
}

func validateHTMLPage(cfg SiteConfig, page PageConfig, markup string) []string {
	var failures []string
	label := page.Name
	if value := elementText(markup, "title"); strings.TrimSpace(value) == "" {
		failures = append(failures, label+": title is empty")
	} else if value != page.Title {
		failures = append(failures, fmt.Sprintf("%s: title is %q", label, value))
	}
	if got := tagAttribute(markup, "meta", "name", "description", "content"); got != page.Description {
		failures = append(failures, label+": meta description is missing or inaccurate")
	}
	canonical := cfg.CanonicalURL(page)
	openGraphImage := cfg.PagesURL + "icons/" + cfg.OpenGraphImage
	checks := []struct {
		tag, identityName, identityValue, targetName, expected, message string
	}{
		{"link", "rel", "canonical", "href", canonical, "self canonical"},
		{"meta", "property", "og:title", "content", page.Title, "Open Graph title"},
		{"meta", "property", "og:description", "content", page.Description, "Open Graph description"},
		{"meta", "property", "og:url", "content", canonical, "Open Graph URL"},
		{"meta", "property", "og:type", "content", "website", "Open Graph type"},
		{"meta", "property", "og:site_name", "content", cfg.DisplayName, "Open Graph site name"},
		{"meta", "property", "og:image", "content", openGraphImage, "Open Graph image"},
		{"meta", "property", "og:image:width", "content", "1200", "Open Graph image width"},
		{"meta", "property", "og:image:height", "content", "630", "Open Graph image height"},
		{"meta", "name", "twitter:card", "content", "summary_large_image", "Twitter card"},
		{"meta", "name", "twitter:image", "content", openGraphImage, "Twitter image"},
		{"link", "rel", "manifest", "href", cfg.AssetURL("manifest.webmanifest"), "manifest link"},
	}
	for _, check := range checks {
		if got := tagAttribute(markup, check.tag, check.identityName, check.identityValue, check.targetName); got != check.expected {
			failures = append(failures, fmt.Sprintf("%s: %s = %q, want %q", label, check.message, got, check.expected))
		}
	}
	if got := len(h1Pattern.FindAllString(markup, -1)); got != 1 {
		failures = append(failures, fmt.Sprintf("%s: h1 count = %d, want 1", label, got))
	}
	if strings.Count(markup, `name="theme-color"`) < 2 {
		failures = append(failures, label+": light and dark theme-color metadata are required")
	}
	visible := html.UnescapeString(tagPattern.ReplaceAllString(scriptPattern.ReplaceAllString(markup, " "), " "))
	if len(strings.Fields(visible)) < 80 {
		failures = append(failures, label+": initial HTML lacks useful visible content")
	}
	for _, item := range []string{"Home", "Examples", "Install", "Usage", "API", "GitHub", "Theme"} {
		if !strings.Contains(markup, item) {
			failures = append(failures, fmt.Sprintf("%s: shared navigation is missing %q", label, item))
		}
	}
	for _, control := range []string{
		`class="theme-picker d-flex flex-nowrap align-items-center gap-2`,
		"<span>Theme</span>",
		`<option value="system">System</option>`,
		`<option value="light">Light</option>`,
		`<option value="dark">Dark</option>`,
	} {
		if !strings.Contains(markup, control) {
			failures = append(failures, label+": shared theme control is incomplete")
		}
	}
	for _, forbidden := range []string{"localhost", "go-package-template", "gopackage", "Formatter", "Labeler", "FormatLabel", "mazey-npm-template", "MAZEY_NPM_TEMPLATE"} {
		if strings.Contains(markup, forbidden) {
			failures = append(failures, fmt.Sprintf("%s: contains forbidden production value %q", label, forbidden))
		}
	}
	if err := validateOwnerLinks(markup, cfg); err != nil {
		failures = append(failures, label+": "+err.Error())
	}
	if !strings.Contains(markup, cfg.Theme.StorageKey) || !strings.Contains(markup, `"system"`) {
		failures = append(failures, label+": generated theme configuration is incomplete")
	}
	if page.Name == "home" || page.Name == "api" {
		expectedImport := fmt.Sprintf(`import "%s"`, cfg.ModulePath)
		if !strings.Contains(markup, expectedImport) {
			failures = append(failures, label+": package import example does not match configuration")
		}
	}
	for _, color := range []string{
		cfg.Theme.Primary.Light.Base, cfg.Theme.Primary.Light.Hover, cfg.Theme.Primary.Light.Active, cfg.Theme.Primary.Light.Soft,
		cfg.Theme.Primary.Dark.Base, cfg.Theme.Primary.Dark.Hover, cfg.Theme.Primary.Dark.Active, cfg.Theme.Primary.Dark.Soft,
	} {
		if !strings.Contains(markup, color) {
			failures = append(failures, fmt.Sprintf("%s: generated theme is missing palette color %s", label, color))
		}
	}
	return failures
}

func validateLinks(outputDir string, cfg SiteConfig, currentPath, markup string) (int, []string) {
	var failures []string
	links := linkPattern.FindAllStringSubmatch(markup, -1)
	for _, match := range links {
		value := html.UnescapeString(match[1])
		if strings.HasPrefix(value, "https://") || strings.HasPrefix(value, "mailto:") || strings.HasPrefix(value, "data:") {
			continue
		}
		if strings.HasPrefix(value, "/") && !strings.HasPrefix(value, cfg.PagesBasePath) {
			failures = append(failures, fmt.Sprintf("%s: unintended root path %q", currentPath, value))
			continue
		}
		pathPart, fragment, _ := strings.Cut(value, "#")
		targetPath := currentPath
		if pathPart != "" {
			if !strings.HasPrefix(pathPart, cfg.PagesBasePath) {
				failures = append(failures, fmt.Sprintf("%s: internal URL is not Pages-subpath-aware: %q", currentPath, value))
				continue
			}
			relative := strings.TrimPrefix(pathPart, cfg.PagesBasePath)
			if relative == "" || strings.HasSuffix(relative, "/") {
				relative += "index.html"
			}
			targetPath = relative
		}
		targetFile := filepath.Join(outputDir, filepath.FromSlash(targetPath))
		info, err := os.Stat(targetFile)
		if err != nil || info.IsDir() {
			failures = append(failures, fmt.Sprintf("%s: broken internal link %q", currentPath, value))
			continue
		}
		if fragment != "" {
			data, err := os.ReadFile(targetFile)
			if err != nil || !regexp.MustCompile(`\bid="`+regexp.QuoteMeta(fragment)+`"`).Match(data) {
				failures = append(failures, fmt.Sprintf("%s: missing anchor for %q", currentPath, value))
			}
		}
	}
	return len(links), failures
}

func validateCrawlerFiles(outputDir string, cfg SiteConfig) []string {
	var failures []string
	robots, err := os.ReadFile(filepath.Join(outputDir, "robots.txt"))
	if err != nil {
		failures = append(failures, "robots.txt is missing")
	} else if string(robots) != Robots(cfg) {
		failures = append(failures, "robots.txt does not match generated policy")
	}
	sitemap, err := os.ReadFile(filepath.Join(outputDir, "sitemap.xml"))
	if err != nil {
		return append(failures, "sitemap.xml is missing")
	}
	matches := locPattern.FindAllStringSubmatch(string(sitemap), -1)
	if len(matches) != 3 {
		failures = append(failures, fmt.Sprintf("sitemap.xml has %d routes, want 3", len(matches)))
	}
	for index, page := range cfg.Pages {
		if index >= len(matches) || matches[index][1] != cfg.CanonicalURL(page) || !strings.HasPrefix(matches[index][1], "https://") {
			failures = append(failures, fmt.Sprintf("sitemap.xml route %d does not match generated page", index+1))
		}
	}
	return failures
}

func validatePWA(outputDir string, cfg SiteConfig, pages map[string]string) (int, []string) {
	var failures []string
	manifestData, err := os.ReadFile(filepath.Join(outputDir, "manifest.webmanifest"))
	if err != nil {
		return 0, []string{"manifest.webmanifest is missing"}
	}
	var manifest webManifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return 0, []string{"manifest.webmanifest is invalid JSON: " + err.Error()}
	}
	if manifest.StartURL != cfg.PagesBasePath || manifest.Scope != cfg.PagesBasePath {
		failures = append(failures, "manifest start_url and scope must match the Pages base path")
	}
	expectedManifestData, manifestErr := Manifest(cfg)
	var expectedManifest webManifest
	if manifestErr != nil || json.Unmarshal(expectedManifestData, &expectedManifest) != nil {
		failures = append(failures, "could not construct the expected web manifest")
	} else if !reflect.DeepEqual(manifest, expectedManifest) {
		failures = append(failures, "manifest.webmanifest does not match configured PWA identity")
	}
	iconCount := 0
	for _, icon := range manifest.Icons {
		relative := strings.TrimPrefix(icon.Src, cfg.PagesBasePath)
		file, err := os.Open(filepath.Join(outputDir, filepath.FromSlash(relative)))
		if err != nil {
			failures = append(failures, "manifest icon is missing: "+icon.Src)
			continue
		}
		imageConfig, _, decodeErr := image.DecodeConfig(file)
		file.Close()
		if decodeErr != nil {
			failures = append(failures, "manifest icon is not a valid image: "+icon.Src)
			continue
		}
		declared := fmt.Sprintf("%dx%d", imageConfig.Width, imageConfig.Height)
		if declared != icon.Sizes {
			failures = append(failures, fmt.Sprintf("manifest icon %s is %s, declared %s", icon.Src, declared, icon.Sizes))
		}
		iconCount++
	}
	configuredImages := []struct {
		name          string
		width, height int
	}{
		{cfg.Favicon, 32, 32},
		{cfg.AppleTouchIcon, 192, 192},
		{cfg.OpenGraphImage, 1200, 630},
	}
	for _, configured := range configuredImages {
		filename := filepath.Join(outputDir, "icons", configured.name)
		file, openErr := os.Open(filename)
		if openErr != nil {
			failures = append(failures, "configured website image is missing: "+configured.name)
			continue
		}
		imageConfig, _, decodeErr := image.DecodeConfig(file)
		file.Close()
		if decodeErr != nil {
			failures = append(failures, "configured website image is invalid: "+configured.name)
			continue
		}
		if imageConfig.Width != configured.width || imageConfig.Height != configured.height {
			failures = append(failures, fmt.Sprintf("configured website image %s is %dx%d, want %dx%d", configured.name, imageConfig.Width, imageConfig.Height, configured.width, configured.height))
		}
	}
	workerData, err := os.ReadFile(filepath.Join(outputDir, "service-worker.js"))
	if err != nil {
		failures = append(failures, "service-worker.js is missing")
		return iconCount, failures
	}
	worker := string(workerData)
	for _, required := range []string{cfg.CachePrefix, cfg.PagesBasePath, cfg.AssetURL("examples/"), cfg.AssetURL("api/"), "url.origin !== self.location.origin", "request.method !== \"GET\"", "MAX_RUNTIME_ENTRIES"} {
		if !strings.Contains(worker, required) {
			failures = append(failures, fmt.Sprintf("service worker is missing %q", required))
		}
	}
	if strings.Contains(worker, "|| url.search") {
		failures = append(failures, "service worker bypasses project requests that contain a query string")
	}
	siteJS, err := os.ReadFile(filepath.Join(outputDir, "assets/js/site.js"))
	if err != nil || !strings.Contains(string(siteJS), "serviceWorker.register") {
		failures = append(failures, "service-worker registration code is missing")
	}
	for path, markup := range pages {
		if !strings.Contains(markup, cfg.AssetURL("service-worker.js")) {
			failures = append(failures, path+": runtime service-worker path is incorrect")
		}
	}
	return iconCount, failures
}

// FindUnresolvedPlaceholders finds template markers and stale identities in generated text files.
func FindUnresolvedPlaceholders(outputDir string) []string {
	forbidden := []string{"TARGET_PROJECT", "PACKAGE_NAME", "MODULE_PATH", "localhost", "go-package-template", "gopackage", "Formatter", "Labeler", "FormatLabel", "mazey-npm-template", "MAZEY_NPM_TEMPLATE"}
	var failures []string
	walkErr := filepath.WalkDir(outputDir, func(filename string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		extension := strings.ToLower(filepath.Ext(filename))
		if extension == ".png" {
			return nil
		}
		data, readErr := os.ReadFile(filename)
		if readErr != nil {
			return readErr
		}
		for _, marker := range forbidden {
			if strings.Contains(string(data), marker) {
				relative, _ := filepath.Rel(outputDir, filename)
				failures = append(failures, fmt.Sprintf("%s contains unresolved value %q", filepath.ToSlash(relative), marker))
			}
		}
		return nil
	})
	if walkErr != nil {
		failures = append(failures, "inspect generated output: "+walkErr.Error())
	}
	return failures
}

func tagAttribute(markup, tag, identityName, identityValue, targetName string) string {
	tags := regexp.MustCompile(`(?is)<`+regexp.QuoteMeta(tag)+`\b[^>]*>`).FindAllString(markup, -1)
	for _, candidate := range tags {
		attributes := make(map[string]string)
		for _, match := range attributePattern.FindAllStringSubmatch(candidate, -1) {
			attributes[strings.ToLower(match[1])] = html.UnescapeString(match[2])
		}
		if strings.EqualFold(attributes[strings.ToLower(identityName)], identityValue) {
			return attributes[strings.ToLower(targetName)]
		}
	}
	return ""
}

func elementText(markup, tag string) string {
	match := regexp.MustCompile(`(?is)<` + regexp.QuoteMeta(tag) + `\b[^>]*>(.*?)</` + regexp.QuoteMeta(tag) + `>`).FindStringSubmatch(markup)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(html.UnescapeString(tagPattern.ReplaceAllString(match[1], "")))
}
