package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type PipScanner struct{}

func (p *PipScanner) Name() Ecosystem { return EcosystemPip }

func (p *PipScanner) Detect(dir string) bool {
	files := []string{"requirements.txt", "pyproject.toml", "setup.py", "Pipfile"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(dir, f)); err == nil {
			return true
		}
	}
	return false
}

type pipOutdated struct {
	Name           string `json:"name"`
	Version        string `json:"version"`
	LatestVersion  string `json:"latest_version"`
}

func (p *PipScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("pip3") && !commandExists("pip") {
		return nil, nil
	}

	pipCmd := "pip3"
	if !commandExists("pip3") {
		pipCmd = "pip"
	}

	out, err := runCommand(dir, pipCmd, "list", "--outdated", "--format=json")
	if err != nil {
		// Try without the outdated flag to at least list installed packages
		out, err = runCommand(dir, pipCmd, "list", "--format=json")
		if err != nil {
			return nil, err
		}
	}

	var deps []Dependency

	var outdated []pipOutdated
	if json.Unmarshal([]byte(out), &outdated) == nil {
		for _, pkg := range outdated {
			dep := Dependency{
				Name:      pkg.Name,
				Current:   pkg.Version,
				Latest:    pkg.LatestVersion,
				Ecosystem: EcosystemPip,
			}
			if dep.Latest == "" {
				dep.Latest = dep.Current
			}
			dep.Outdated = dep.Current != dep.Latest
			deps = append(deps, dep)
		}
	}

	// Also parse requirements.txt for context
	reqPath := filepath.Join(dir, "requirements.txt")
	if data, err := os.ReadFile(reqPath); err == nil {
		existing := make(map[string]bool)
		for _, d := range deps {
			existing[strings.ToLower(d.Name)] = true
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "==", 2)
			name := strings.TrimSpace(parts[0])
			// Strip version specifiers like >=, <=, ~=
			for _, sep := range []string{">=", "<=", "~=", "!=", ">"} {
				if idx := strings.Index(name, sep); idx != -1 {
					name = name[:idx]
				}
			}
			name = strings.TrimSpace(name)
			if name != "" && !existing[strings.ToLower(name)] {
				dep := Dependency{
					Name:      name,
					Ecosystem: EcosystemPip,
				}
				if len(parts) == 2 {
					dep.Current = strings.TrimSpace(parts[1])
					dep.Latest = dep.Current
				}
				deps = append(deps, dep)
			}
		}
	}

	return deps, nil
}

func (p *PipScanner) Update(dir string, dep Dependency) error {
	pipCmd := "pip3"
	if !commandExists("pip3") {
		pipCmd = "pip"
	}
	_, err := runCommand(dir, pipCmd, "install", "--upgrade", dep.Name)
	return err
}
