// Command sitegen builds, validates, and previews the generated GitHub Pages website.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chengchuu/gurl/internal/site"
)

func main() {
	log.SetFlags(0)
	workingDirectory, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root, err := site.FindProjectRoot(workingDirectory)
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := site.LoadConfig(root)
	if err != nil {
		log.Fatal(err)
	}
	command := "build"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	switch command {
	case "build":
		result, err := site.Build(root, cfg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("generated %d pages and %d precached routes/assets in %s (version %s)\n", result.Pages, result.Assets, result.OutputDir, result.Version)
	case "validate":
		result, err := site.ValidateArtifact(root, cfg)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("validated %d pages, %d links, and %d PWA icons in %s\n", result.Pages, result.Links, result.Icons, result.OutputDir)
	case "preview":
		address := "127.0.0.1:4173"
		if value := os.Getenv("SITEGEN_ADDR"); value != "" {
			address = value
		}
		if err := site.Preview(root, cfg, address); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unknown command %q; use build, validate, or preview", command)
	}
}
