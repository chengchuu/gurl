package site

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// SiteConfig is the build-time source of truth for project, website, SEO, theme, and PWA identity.
// It belongs to internal tooling and is not part of the public package API.
type SiteConfig struct {
	PackageName     string
	DisplayName     string
	ModulePath      string
	Description     string
	RepositoryOwner string
	RepositoryName  string
	RepositoryURL   string
	PagesBasePath   string
	PagesURL        string
	Theme           ThemeConfig
	CachePrefix     string
	PWAShortName    string
	GoVersion       string
	License         string
	OutputDir       string
	SiteSourceDir   string
	Favicon         string
	Icon192         string
	Icon512         string
	MaskableIcon512 string
	AppleTouchIcon  string
	OpenGraphImage  string
	Pages           []PageConfig
}

// ThemeConfig owns the website color modes and persisted preference identity.
type ThemeConfig struct {
	StorageKey string       `json:"-"`
	ColorLight string       `json:"colorLight"`
	ColorDark  string       `json:"colorDark"`
	Primary    ThemePalette `json:"primary"`
}

// ThemePalette contains the coordinated primary colors for both Bootstrap modes.
type ThemePalette struct {
	Light PrimaryPalette `json:"light"`
	Dark  PrimaryPalette `json:"dark"`
}

// PrimaryPalette defines interactive and soft-accent colors for one mode.
type PrimaryPalette struct {
	Base      string `json:"base"`
	Hover     string `json:"hover"`
	Active    string `json:"active"`
	Soft      string `json:"soft"`
	RGB       string `json:"rgb"`
	HoverRGB  string `json:"hoverRgb"`
	ActiveRGB string `json:"activeRgb"`
}

// PageConfig owns the unique identity and production location of one public page.
type PageConfig struct {
	Name        string
	OutputPath  string
	Path        string
	Title       string
	Description string
	SchemaType  string
}

// pageConfigFile contains the configured identity for one public page.
type pageConfigFile struct {
	Name        string `json:"name"`
	OutputPath  string `json:"outputPath"`
	Path        string `json:"path"`
	Title       string `json:"title"`
	Description string `json:"description"`
	SchemaType  string `json:"schemaType"`
}

type configFile struct {
	PackageName string           `json:"packageName"`
	DisplayName string           `json:"displayName"`
	ModulePath  string           `json:"modulePath"`
	Description string           `json:"description"`
	Repository  repositoryConfig `json:"repository"`
	PWA         pwaConfig        `json:"pwa"`
	GoVersion   string           `json:"goVersion"`
	License     string           `json:"license"`
	Icons       iconConfig       `json:"icons"`
	Theme       ThemeConfig      `json:"theme"`
	Pages       []pageConfigFile `json:"pages"`
}

type repositoryConfig struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type pwaConfig struct {
	ShortName string `json:"shortName"`
}

type iconConfig struct {
	Favicon         string `json:"favicon"`
	Icon192         string `json:"icon192"`
	Icon512         string `json:"icon512"`
	MaskableIcon512 string `json:"maskableIcon512"`
	AppleTouchIcon  string `json:"appleTouchIcon"`
	OpenGraphImage  string `json:"openGraphImage"`
}

type stablePage struct {
	OutputPath string
	Path       string
	SchemaType string
}

type moduleDefinition struct {
	Path      string
	GoVersion string
}

var (
	githubOwnerPattern      = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,37}[A-Za-z0-9])?$`)
	githubRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,100}$`)
	stablePages             = map[string]stablePage{
		"home":     {OutputPath: "index.html", Path: "", SchemaType: "SoftwareSourceCode"},
		"examples": {OutputPath: "examples/index.html", Path: "examples/", SchemaType: "TechArticle"},
		"api":      {OutputPath: "api/index.html", Path: "api/", SchemaType: "TechArticle"},
	}
)

// LoadConfig reads the editable site.config.json file and derives deployment-specific values.
func LoadConfig(projectRoot string) (SiteConfig, error) {
	filename := filepath.Join(projectRoot, "site.config.json")
	file, err := os.Open(filename)
	if err != nil {
		return SiteConfig{}, fmt.Errorf("open site configuration: %w", err)
	}
	defer file.Close()

	var source configFile
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&source); err != nil {
		return SiteConfig{}, fmt.Errorf("decode site configuration: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return SiteConfig{}, fmt.Errorf("decode site configuration: expected one JSON object")
	}
	module, err := readModuleDefinition(projectRoot)
	if err != nil {
		return SiteConfig{}, err
	}
	if source.ModulePath != module.Path {
		return SiteConfig{}, fmt.Errorf("site module path %q does not match go.mod module path %q", source.ModulePath, module.Path)
	}
	if source.GoVersion != module.GoVersion && source.GoVersion != strings.TrimSuffix(module.GoVersion, ".0") {
		return SiteConfig{}, fmt.Errorf("site Go version %q does not match go.mod Go version %q", source.GoVersion, module.GoVersion)
	}

	repositoryURL := repositoryURL(source.Repository.Owner, source.Repository.Name)
	basePath := DerivePagesBasePath(source.Repository.Name)
	theme := source.Theme
	theme.StorageKey = strings.TrimSpace(source.Repository.Name) + "-theme"
	pages := make([]PageConfig, 0, len(source.Pages))
	for _, page := range source.Pages {
		pages = append(pages, PageConfig{
			Name:        page.Name,
			OutputPath:  page.OutputPath,
			Path:        page.Path,
			Title:       page.Title,
			Description: page.Description,
			SchemaType:  page.SchemaType,
		})
	}
	cfg := SiteConfig{
		PackageName:     source.PackageName,
		DisplayName:     source.DisplayName,
		ModulePath:      source.ModulePath,
		Description:     source.Description,
		RepositoryOwner: source.Repository.Owner,
		RepositoryName:  source.Repository.Name,
		RepositoryURL:   repositoryURL,
		PagesBasePath:   basePath,
		PagesURL:        ProductionURL(source.Repository.Owner, source.Repository.Name),
		Theme:           theme,
		CachePrefix:     strings.TrimSpace(source.Repository.Name) + "-site-",
		PWAShortName:    source.PWA.ShortName,
		GoVersion:       source.GoVersion,
		License:         source.License,
		OutputDir:       "dist/pages",
		SiteSourceDir:   "site",
		Favicon:         source.Icons.Favicon,
		Icon192:         source.Icons.Icon192,
		Icon512:         source.Icons.Icon512,
		MaskableIcon512: source.Icons.MaskableIcon512,
		AppleTouchIcon:  source.Icons.AppleTouchIcon,
		OpenGraphImage:  source.Icons.OpenGraphImage,
		Pages:           pages,
	}
	if err := cfg.Validate(); err != nil {
		return SiteConfig{}, fmt.Errorf("validate site configuration: %w", err)
	}
	return cfg, nil
}

func readModuleDefinition(projectRoot string) (moduleDefinition, error) {
	filename := filepath.Join(projectRoot, "go.mod")
	data, err := os.ReadFile(filename)
	if err != nil {
		return moduleDefinition{}, fmt.Errorf("read module definition: %w", err)
	}
	var result moduleDefinition
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "module":
			result.Path = fields[1]
			if strings.HasPrefix(result.Path, `"`) || strings.HasPrefix(result.Path, "`") {
				result.Path, err = strconv.Unquote(result.Path)
				if err != nil {
					return moduleDefinition{}, fmt.Errorf("parse module path in %s: %w", filename, err)
				}
			}
		case "go":
			result.GoVersion = fields[1]
		}
	}
	if result.Path == "" {
		return moduleDefinition{}, fmt.Errorf("module directive not found in %s", filename)
	}
	if result.GoVersion == "" {
		return moduleDefinition{}, fmt.Errorf("Go version directive not found in %s", filename)
	}
	return result, nil
}

func repositoryURL(owner, repositoryName string) string {
	return fmt.Sprintf("https://github.com/%s/%s", strings.TrimSpace(owner), strings.TrimSpace(repositoryName))
}

// DerivePagesBasePath derives a GitHub project Pages path from a repository name.
func DerivePagesBasePath(repositoryName string) string {
	name := strings.Trim(strings.TrimSpace(repositoryName), "/")
	if name == "" || strings.Contains(name, "..") {
		return "/"
	}
	return "/" + url.PathEscape(name) + "/"
}

// ProductionURL derives the canonical GitHub Pages URL for a project repository.
func ProductionURL(owner, repositoryName string) string {
	owner = strings.TrimSpace(owner)
	base := DerivePagesBasePath(repositoryName)
	return fmt.Sprintf("https://%s.github.io%s", owner, base)
}

// CanonicalURL returns the absolute production URL for page.
func (c SiteConfig) CanonicalURL(page PageConfig) string {
	return c.PagesURL + page.Path
}

// AssetURL returns a repository-subpath-aware URL for a generated asset.
func (c SiteConfig) AssetURL(assetPath string) string {
	cleaned := strings.TrimPrefix(path.Clean("/"+assetPath), "/")
	if strings.HasSuffix(assetPath, "/") && cleaned != "" {
		cleaned += "/"
	}
	return c.PagesBasePath + cleaned
}

// Page finds a configured page by name.
func (c SiteConfig) Page(name string) (PageConfig, bool) {
	for _, page := range c.Pages {
		if page.Name == name {
			return page, true
		}
	}
	return PageConfig{}, false
}

// Validate checks invariants needed by generation and deployment.
func (c SiteConfig) Validate() error {
	required := []struct{ label, value string }{
		{"package name", c.PackageName}, {"display name", c.DisplayName}, {"module path", c.ModulePath},
		{"description", c.Description}, {"repository owner", c.RepositoryOwner}, {"repository name", c.RepositoryName},
		{"repository URL", c.RepositoryURL}, {"Pages URL", c.PagesURL}, {"theme storage key", c.Theme.StorageKey},
		{"cache prefix", c.CachePrefix}, {"PWA short name", c.PWAShortName}, {"Go version", c.GoVersion},
		{"license", c.License}, {"favicon", c.Favicon}, {"192px icon", c.Icon192}, {"512px icon", c.Icon512},
		{"maskable icon", c.MaskableIcon512}, {"Apple touch icon", c.AppleTouchIcon}, {"Open Graph image", c.OpenGraphImage},
		{"light background", c.Theme.ColorLight}, {"dark background", c.Theme.ColorDark},
		{"light primary", c.Theme.Primary.Light.Base}, {"light primary hover", c.Theme.Primary.Light.Hover},
		{"light primary active", c.Theme.Primary.Light.Active}, {"light primary soft", c.Theme.Primary.Light.Soft},
		{"light primary RGB", c.Theme.Primary.Light.RGB}, {"light primary hover RGB", c.Theme.Primary.Light.HoverRGB},
		{"light primary active RGB", c.Theme.Primary.Light.ActiveRGB},
		{"dark primary", c.Theme.Primary.Dark.Base}, {"dark primary hover", c.Theme.Primary.Dark.Hover},
		{"dark primary active", c.Theme.Primary.Dark.Active}, {"dark primary soft", c.Theme.Primary.Dark.Soft},
		{"dark primary RGB", c.Theme.Primary.Dark.RGB}, {"dark primary hover RGB", c.Theme.Primary.Dark.HoverRGB},
		{"dark primary active RGB", c.Theme.Primary.Dark.ActiveRGB},
	}
	for _, item := range required {
		if strings.TrimSpace(item.value) == "" {
			return fmt.Errorf("%s must not be empty", item.label)
		}
	}
	images := []struct{ label, filename string }{
		{"favicon", c.Favicon}, {"192px icon", c.Icon192}, {"512px icon", c.Icon512},
		{"maskable icon", c.MaskableIcon512}, {"Apple touch icon", c.AppleTouchIcon},
		{"Open Graph image", c.OpenGraphImage},
	}
	for _, image := range images {
		if strings.ContainsAny(image.filename, `/\`) || !strings.EqualFold(filepath.Ext(image.filename), ".png") {
			return fmt.Errorf("%s %q must be a PNG filename without a path", image.label, image.filename)
		}
	}
	if !githubOwnerPattern.MatchString(c.RepositoryOwner) {
		return fmt.Errorf("repository owner %q is not a valid GitHub owner", c.RepositoryOwner)
	}
	if !githubRepositoryPattern.MatchString(c.RepositoryName) || c.RepositoryName == "." || c.RepositoryName == ".." {
		return fmt.Errorf("repository name %q is not a valid GitHub repository name", c.RepositoryName)
	}
	if c.RepositoryURL != repositoryURL(c.RepositoryOwner, c.RepositoryName) {
		return fmt.Errorf("repository URL %q is not derived from repository identity", c.RepositoryURL)
	}
	if c.PagesBasePath != DerivePagesBasePath(c.RepositoryName) {
		return fmt.Errorf("Pages base path %q does not match repository %q", c.PagesBasePath, c.RepositoryName)
	}
	if c.PagesURL != ProductionURL(c.RepositoryOwner, c.RepositoryName) {
		return fmt.Errorf("Pages URL %q is not derived from repository identity", c.PagesURL)
	}
	if c.Theme.StorageKey != c.RepositoryName+"-theme" {
		return fmt.Errorf("theme storage key %q is not derived from repository identity", c.Theme.StorageKey)
	}
	if c.CachePrefix != c.RepositoryName+"-site-" {
		return fmt.Errorf("cache prefix %q is not derived from repository identity", c.CachePrefix)
	}
	if len(c.Pages) != len(stablePages) {
		return fmt.Errorf("exactly three stable public pages are required")
	}
	seenNames := make(map[string]bool)
	seenTitles := make(map[string]bool)
	seenDescriptions := make(map[string]bool)
	for _, page := range c.Pages {
		if page.Name == "" || page.OutputPath == "" || strings.TrimSpace(page.Title) == "" || strings.TrimSpace(page.Description) == "" || page.SchemaType == "" {
			return fmt.Errorf("page configuration is incomplete: %+v", page)
		}
		if page.Title != strings.TrimSpace(page.Title) {
			return fmt.Errorf("page %q title must not have surrounding whitespace", page.Name)
		}
		if seenNames[page.Name] {
			return fmt.Errorf("duplicate page name %q", page.Name)
		}
		seenNames[page.Name] = true
		expected, ok := stablePages[page.Name]
		if !ok {
			return fmt.Errorf("page %q is not one of the stable public pages", page.Name)
		}
		if page.OutputPath != expected.OutputPath || page.Path != expected.Path || page.SchemaType != expected.SchemaType {
			return fmt.Errorf("page %q route or schema does not match the stable public contract", page.Name)
		}
		if seenTitles[page.Title] {
			return fmt.Errorf("duplicate page title %q", page.Title)
		}
		seenTitles[page.Title] = true
		if seenDescriptions[page.Description] {
			return fmt.Errorf("duplicate page description %q", page.Description)
		}
		seenDescriptions[page.Description] = true
	}
	return nil
}
