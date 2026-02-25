package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type NpmScanner struct{}

func (n *NpmScanner) Name() Ecosystem { return EcosystemNpm }

func (n *NpmScanner) Detect(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "package.json")); err != nil {
		return false
	}
	// Skip if this is a Bun project (has bun.lockb or bun.lock)
	if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "bun.lock")); err == nil {
		return false
	}
	return true
}

type npmOutdated struct {
	Current string `json:"current"`
	Wanted  string `json:"wanted"`
	Latest  string `json:"latest"`
}

func (n *NpmScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("npm") {
		return nil, nil
	}

	out, _ := runCommand(dir, "npm", "outdated", "--json")
	// npm outdated exits non-zero when there ARE outdated packages

	var deps []Dependency
	outdatedMap := make(map[string]bool)

	if out != "" {
		var outdated map[string]npmOutdated
		if err := json.Unmarshal([]byte(out), &outdated); err == nil {
			for name, info := range outdated {
				current := info.Current
				// When node_modules is absent, npm omits "current"; use "wanted" instead
				if current == "" {
					current = info.Wanted
				}
				dep := Dependency{
					Name:      name,
					Current:   current,
					Latest:    info.Latest,
					Ecosystem: EcosystemNpm,
					Outdated:  current != info.Latest,
				}
				deps = append(deps, dep)
				outdatedMap[name] = true
			}
		}
	}

	// Also parse package.json for all deps (including up-to-date)
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err == nil {
		var pkg struct {
			Dependencies    map[string]string `json:"dependencies"`
			DevDependencies map[string]string `json:"devDependencies"`
		}
		if json.Unmarshal(data, &pkg) == nil {
			// Collect packages not covered by npm outdated
			var uncovered []struct {
				name string
				ver  string
				dev  bool
			}
			for name, ver := range pkg.Dependencies {
				if !outdatedMap[name] {
					uncovered = append(uncovered, struct {
						name string
						ver  string
						dev  bool
					}{name, ver, false})
				}
			}
			for name, ver := range pkg.DevDependencies {
				if !outdatedMap[name] {
					uncovered = append(uncovered, struct {
						name string
						ver  string
						dev  bool
					}{name, ver, true})
				}
			}

			// Look up latest versions in parallel
			type result struct {
				idx    int
				latest string
			}
			results := make([]result, len(uncovered))
			var wg sync.WaitGroup
			for i, pkg := range uncovered {
				wg.Add(1)
				go func(idx int, pkgName string) {
					defer wg.Done()
					latest, err := runCommand(dir, "npm", "view", pkgName, "version")
					if err == nil && latest != "" {
						results[idx] = result{idx: idx, latest: latest}
					}
				}(i, pkg.name)
			}
			wg.Wait()

			for i, pkg := range uncovered {
				current := cleanVersion(pkg.ver)
				latest := current
				if results[i].latest != "" {
					latest = results[i].latest
				}
				deps = append(deps, Dependency{
					Name:      pkg.name,
					Current:   current,
					Latest:    latest,
					Ecosystem: EcosystemNpm,
					Outdated:  current != latest,
					Indirect:  pkg.dev,
				})
			}
		}
	}

	// Check for vulnerabilities
	n.checkVulns(dir, deps)

	return deps, nil
}

// cleanVersion strips semver range prefixes (^, ~, >=, etc.) from a version string.
func cleanVersion(v string) string {
	v = strings.TrimSpace(v)
	for len(v) > 0 && (v[0] == '^' || v[0] == '~' || v[0] == '>' || v[0] == '=' || v[0] == '<') {
		v = v[1:]
	}
	return strings.TrimSpace(v)
}

func (n *NpmScanner) checkVulns(dir string, deps []Dependency) {
	out, err := runCommand(dir, "npm", "audit", "--json")
	if err != nil || out == "" {
		return
	}
	var audit struct {
		Vulnerabilities map[string]struct {
			Severity string `json:"severity"`
		} `json:"vulnerabilities"`
	}
	if json.Unmarshal([]byte(out), &audit) == nil {
		for i, dep := range deps {
			if v, ok := audit.Vulnerabilities[dep.Name]; ok {
				deps[i].Vulnerable = true
				deps[i].VulnInfo = "npm audit: " + v.Severity
			}
		}
	}
}

func (n *NpmScanner) Update(dir string, dep Dependency) error {
	_, err := runCommand(dir, "npm", "install", dep.Name+"@"+dep.Latest)
	return err
}
