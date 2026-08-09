package site

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	githubRepositoryLinkPattern = regexp.MustCompile(`https://github\.com/([A-Za-z0-9-]+)/([A-Za-z0-9._-]+)`)
	githubPagesLinkPattern      = regexp.MustCompile(`https://([A-Za-z0-9-]+)\.github\.io/([A-Za-z0-9._-]+)/?`)
)

// ValidateSourceIdentity checks the maintained README mirrors of derived project identity.
func ValidateSourceIdentity(projectRoot string, cfg SiteConfig) error {
	data, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
	if err != nil {
		return fmt.Errorf("read README identity: %w", err)
	}
	readme := string(data)
	required := []string{
		"go get " + cfg.ModulePath,
		`import "` + cfg.ModulePath + `"`,
		cfg.RepositoryURL,
		cfg.PagesURL,
		cfg.PagesURL + "examples/",
		cfg.PagesURL + "api/",
		"https://pkg.go.dev/" + cfg.ModulePath,
	}
	for _, value := range required {
		if !strings.Contains(readme, value) {
			return fmt.Errorf("README is missing configured identity %q", value)
		}
	}
	if err := validateOwnerLinks(readme, cfg); err != nil {
		return fmt.Errorf("README identity: %w", err)
	}
	return nil
}

func validateOwnerLinks(content string, cfg SiteConfig) error {
	for _, match := range githubRepositoryLinkPattern.FindAllStringSubmatch(content, -1) {
		if match[1] != cfg.RepositoryOwner || match[2] != cfg.RepositoryName {
			return fmt.Errorf("repository URL %q does not match configured owner and project", match[0])
		}
	}
	for _, match := range githubPagesLinkPattern.FindAllStringSubmatch(content, -1) {
		if match[1] != cfg.RepositoryOwner || match[2] != cfg.RepositoryName {
			return fmt.Errorf("Pages URL %q does not match configured owner and project", match[0])
		}
	}
	return nil
}
