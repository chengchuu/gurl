package site

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Preview builds, validates, and serves the production-like artifact below its Pages base path.
func Preview(projectRoot string, cfg SiteConfig, address string) error {
	if _, err := Build(projectRoot, cfg); err != nil {
		return err
	}
	if _, err := ValidateArtifact(projectRoot, cfg); err != nil {
		return err
	}
	outputDir := filepath.Join(projectRoot, cfg.OutputDir)
	baseWithoutSlash := strings.TrimSuffix(cfg.PagesBasePath, "/")
	files := http.FileServer(http.Dir(outputDir))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, cfg.PagesBasePath, http.StatusTemporaryRedirect)
	})
	mux.Handle(baseWithoutSlash+"/", http.StripPrefix(baseWithoutSlash, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasSuffix(request.URL.Path, "service-worker.js") {
			writer.Header().Set("Cache-Control", "no-cache")
			writer.Header().Set("Service-Worker-Allowed", cfg.PagesBasePath)
		}
		if strings.HasSuffix(request.URL.Path, ".webmanifest") {
			writer.Header().Set("Content-Type", "application/manifest+json")
		}
		files.ServeHTTP(writer, request)
	})))
	log.Printf("production-like Pages preview: http://%s%s", address, cfg.PagesBasePath)
	return http.ListenAndServe(address, mux)
}

// FindProjectRoot searches upward for the module definition.
func FindProjectRoot(start string) (string, error) {
	current, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if info, statErr := os.Stat(filepath.Join(current, "go.mod")); statErr == nil && !info.IsDir() {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		current = parent
	}
}
