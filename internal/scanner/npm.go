package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	if out != "" {
		var outdated map[string]npmOutdated
		if err := json.Unmarshal([]byte(out), &outdated); err == nil {
			for name, info := range outdated {
				dep := Dependency{
					Name:      name,
					Current:   info.Current,
					Latest:    info.Latest,
					Ecosystem: EcosystemNpm,
					Outdated:  info.Current != info.Latest,
				}
				deps = append(deps, dep)
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
			existing := make(map[string]bool)
			for _, d := range deps {
				existing[d.Name] = true
			}
			for name, ver := range pkg.Dependencies {
				if !existing[name] {
					deps = append(deps, Dependency{
						Name:      name,
						Current:   ver,
						Latest:    ver,
						Ecosystem: EcosystemNpm,
					})
				}
			}
			for name, ver := range pkg.DevDependencies {
				if !existing[name] {
					deps = append(deps, Dependency{
						Name:      name,
						Current:   ver,
						Latest:    ver,
						Ecosystem: EcosystemNpm,
						Indirect:  true,
					})
				}
			}
		}
	}

	// Check for vulnerabilities
	n.checkVulns(dir, deps)

	return deps, nil
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
