package scanner

import (
	"os"
	"path/filepath"
	"strings"
)

type CargoScanner struct{}

func (c *CargoScanner) Name() Ecosystem { return EcosystemCargo }

func (c *CargoScanner) Detect(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "Cargo.toml"))
	return err == nil
}

func (c *CargoScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("cargo") {
		return nil, nil
	}

	out, err := runCommand(dir, "cargo", "outdated", "--root-deps-only", "--format=list")
	if err != nil {
		// cargo-outdated might not be installed, fall back to Cargo.toml parsing
		return c.parseCargo(dir)
	}

	var deps []Dependency
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Name") || strings.HasPrefix(line, "---") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			dep := Dependency{
				Name:      fields[0],
				Current:   fields[1],
				Ecosystem: EcosystemCargo,
			}
			// Format: Name Project Compat Latest Kind
			if len(fields) >= 4 {
				dep.Latest = fields[3]
			} else {
				dep.Latest = fields[2]
			}
			dep.Outdated = dep.Current != dep.Latest
			deps = append(deps, dep)
		}
	}

	if len(deps) == 0 {
		return c.parseCargo(dir)
	}

	// Check for vulnerabilities
	c.checkVulns(dir, deps)

	return deps, nil
}

func (c *CargoScanner) parseCargo(dir string) ([]Dependency, error) {
	data, err := os.ReadFile(filepath.Join(dir, "Cargo.toml"))
	if err != nil {
		return nil, err
	}

	var deps []Dependency
	inDeps := false

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "[dependencies]" || line == "[dev-dependencies]" {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}
		if inDeps && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			name := strings.TrimSpace(parts[0])
			version := strings.Trim(strings.TrimSpace(parts[1]), "\"")
			if name != "" {
				deps = append(deps, Dependency{
					Name:      name,
					Current:   version,
					Latest:    version,
					Ecosystem: EcosystemCargo,
				})
			}
		}
	}

	return deps, nil
}

func (c *CargoScanner) checkVulns(dir string, deps []Dependency) {
	if !commandExists("cargo-audit") {
		return
	}
	out, err := runCommand(dir, "cargo", "audit", "--json")
	if err != nil || out == "" {
		return
	}
	for i, dep := range deps {
		if strings.Contains(out, dep.Name) {
			deps[i].Vulnerable = true
			deps[i].VulnInfo = "Reported by cargo-audit"
		}
	}
}

func (c *CargoScanner) Update(dir string, dep Dependency) error {
	_, err := runCommand(dir, "cargo", "update", "-p", dep.Name)
	return err
}
