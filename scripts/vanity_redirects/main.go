// SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/go-github/v68/github"
	"golang.org/x/mod/modfile"
)

type tmplData struct {
	Module string
	Repo   string
}

func generate(mod, repo, staticDir, prefix string) error {
	tmpl, err := template.New("index").Parse(`<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en" lang="en-us">
  <head>
    <meta http-equiv="content-type" content="text/html; charset=utf-8">
    <meta name="go-import" content="{{ .Module }} git {{ .Repo }}">
    <meta http-equiv="Refresh" content="0; url={{ .Repo }}" />
  </head>
  <body>
  </body>
</html>`)
	if err != nil {
		return err
	}

	subDir := strings.TrimPrefix(mod, prefix)

	dir := filepath.Join(staticDir, subDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	fn := filepath.Join(dir, "index.html")

	f, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(f, tmplData{
		Module: mod,
		Repo:   repo,
	}); err != nil {
		return err
	}

	if _, err := f.WriteString(`
<!--
  SPDX-FileCopyrightText: 2023-2025 Steffen Vogel <post@steffenvogel.de>
  SPDX-License-Identifier: Apache-2.0
-->
  `); err != nil {
		return err
	}

	log.Println("Added redirect", mod, repo)

	return f.Close()
}

func getModsFromGitHub(ctx context.Context, owner string) (map[string]string, error) {
	client := github.NewClient(nil)

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		client = client.WithAuthToken(token)
	}

	repos, _, err := client.Repositories.ListByUser(ctx, owner, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repos: %w", err)
	}

	mods := map[string]string{}

	for _, repo := range repos {
		file, _, err := client.Repositories.DownloadContents(ctx, owner, *repo.Name, "go.mod", nil)
		if err != nil {
			log.Printf("Failed to download %s/%s/go.mod: %v", owner, *repo.Name, err)

			continue
		}

		defer file.Close()

		contents, err := io.ReadAll(file)
		if err != nil {
			return nil, fmt.Errorf("failed to download go.mod: %w", err)
		}

		modFile, err := modfile.Parse("go.mod", contents, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to parse go.mod: %w", err)
		}

		if modFile.Module == nil {
			continue
		}

		mods[modFile.Module.Mod.Path] = *repo.CloneURL
	}

	return mods, nil
}

func getModBase(pkg string) string {
	re := regexp.MustCompile(`\/v[0-9]`)

	return re.ReplaceAllString(pkg, "")
}

func main() {
	prefix := flag.String("prefix", "cunicu.li", "Prefix")
	owner := flag.String("owner", "cunicu", "GitHub user/org")
	modsFile := flag.String("modules-file", "", "JSON file containing a map of modules")
	modsFetch := flag.Bool("modules-fetch", false, "Only fetch modules from GitHub")
	staticDir := flag.String("static-dir", "", "Directory in which the generated files should be placed")
	flag.Parse()

	ctx := context.Background()

	var mods map[string]string

	if *modsFile == "" || *modsFetch {
		var err error
		if mods, err = getModsFromGitHub(ctx, *owner); err != nil {
			log.Fatalf("Failed to list packages: %v", err)
		}
	}

	if *modsFetch {
		if *modsFile == "" {
			log.Fatal("-modules-fetch requires -modules-file")
		}

		modsJSON, err := json.MarshalIndent(mods, "", "  ")
		if err != nil {
			log.Fatalf("Failed to encode modules: %s", err)
		}

		if err := os.WriteFile(*modsFile, modsJSON, 0o755); err != nil { //nolint:gosec
			log.Fatalf("Failed to write modules to file: %s", err)
		}
	} else if *modsFile != "" {
		modsJSON, err := os.ReadFile(*modsFile)
		if err != nil {
			log.Fatalf("Failed to read modules from file: %s", err)
		}

		if err := json.Unmarshal(modsJSON, &mods); err != nil {
			log.Fatalf("Failed to decode modules: %s", err)
		}
	}

	if *staticDir != "" {
		for mod, repo := range mods {
			modBase := getModBase(mod)

			if !strings.HasPrefix(mod, *prefix) {
				continue
			}

			if err := generate(mod, repo, *staticDir, *prefix); err != nil {
				log.Fatalf("Failed to generate HTML file: %v", err)
			}

			if mod == modBase {
				continue
			}

			if err := generate(modBase, repo, *staticDir, *prefix); err != nil {
				log.Fatalf("Failed to generate HTML file: %v", err)
			}
		}
	}
}
