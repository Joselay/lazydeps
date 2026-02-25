package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type GoScanner struct{}

func (g *GoScanner) Name() Ecosystem { return EcosystemGo }

func (g *GoScanner) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

type goListModule struct {
	Path    string
	Version string
	Update  *struct {
		Version string
	}
	Indirect bool
}

func (g *GoScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("go") {
		return nil, nil
	}

	out, err := runCommand(dir, "go", "list", "-m", "-u", "-json", "all")
	if err != nil {
		return nil, err
	}

	var deps []Dependency
	decoder := json.NewDecoder(strings.NewReader(out))

	for decoder.More() {
		var mod goListModule
		if err := decoder.Decode(&mod); err != nil {
			continue
		}
		// Skip the main module
		if mod.Version == "" {
			continue
		}

		dep := Dependency{
			Name:      mod.Path,
			Current:   mod.Version,
			Ecosystem: EcosystemGo,
			Indirect:  mod.Indirect,
		}

		if mod.Update != nil {
			dep.Latest = mod.Update.Version
			dep.Outdated = true
		} else {
			dep.Latest = mod.Version
		}

		deps = append(deps, dep)
	}

	// Check for vulnerabilities
	g.checkVulns(dir, deps)

	return deps, nil
}

func (g *GoScanner) checkVulns(dir string, deps []Dependency) {
	if !commandExists("govulncheck") {
		return
	}
	out, err := runCommand(dir, "govulncheck", "-json", "./...")
	if err != nil || out == "" {
		return
	}
	// Simple heuristic: mark deps whose names appear in govulncheck output
	for i, dep := range deps {
		if strings.Contains(out, dep.Name) {
			deps[i].Vulnerable = true
			deps[i].VulnInfo = "Potential vulnerability detected by govulncheck"
		}
	}
}

func (g *GoScanner) Update(dir string, dep Dependency) error {
	target := dep.Name + "@" + dep.Latest
	if _, err := runCommand(dir, "go", "get", target); err != nil {
		return err
	}
	_, err := runCommand(dir, "go", "mod", "tidy")
	return err
}
