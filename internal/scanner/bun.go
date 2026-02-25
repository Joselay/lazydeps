package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type BunScanner struct{}

func (b *BunScanner) Name() Ecosystem { return EcosystemBun }

func (b *BunScanner) Detect(dir string) bool {
	// Bun project: has bun.lockb (or bun.lock) alongside package.json
	if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(dir, "bun.lock")); err == nil {
		return true
	}
	return false
}

type bunOutdated struct {
	Name    string `json:"name"`
	Current string `json:"current"`
	Update  string `json:"update"`
	Latest  string `json:"latest"`
}

func (b *BunScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("bun") {
		return nil, nil
	}

	deps := b.scanOutdated(dir)

	// Fall back to package.json parsing if nothing returned
	if len(deps) == 0 {
		deps = b.fallbackPackageJSON(dir)
	}

	return deps, nil
}

func (b *BunScanner) scanOutdated(dir string) []Dependency {
	// bun outdated (available since bun 1.1.28) outputs text table
	// bun outdated exits non-zero when there ARE outdated packages, so ignore the error
	out, _ := runCommand(dir, "bun", "outdated")
	if out == "" {
		return nil
	}

	// Parse the text table output. Bun uses pipe characters:
	// |------------------------------------------------------------|
	// | Package                    | Current  | Update   | Latest  |
	// |----------------------------|----------|----------|---------|
	// | @ai-sdk/react              | 3.0.99   | 3.0.101  | 3.0.101 |
	// | @tailwindcss/postcss (dev) | 4.1.18   | 4.2.1    | 4.2.1   |
	// |------------------------------------------------------------|
	var deps []Dependency
	lines := splitLines(out)
	for _, line := range lines {
		// Skip separator rows like |------|------|
		if isSeparatorRow(line) {
			continue
		}
		fields := parseTableRow(line)
		if len(fields) < 4 {
			continue
		}
		name := fields[0]
		// Skip header row
		if name == "Package" || name == "" {
			continue
		}

		// Handle "(dev)" suffix — marks dev dependencies
		isDev := false
		if strings.HasSuffix(name, "(dev)") {
			name = strings.TrimSpace(strings.TrimSuffix(name, "(dev)"))
			isDev = true
		}

		current := fields[1]
		latest := fields[3]
		if latest == "" {
			latest = fields[2]
		}

		dep := Dependency{
			Name:      name,
			Current:   current,
			Latest:    latest,
			Ecosystem: EcosystemBun,
			Outdated:  current != latest,
			Indirect:  isDev,
		}
		deps = append(deps, dep)
	}

	// Also add deps from package.json not covered by bun outdated,
	// looking up their real latest version from the registry
	existing := make(map[string]bool)
	for _, d := range deps {
		existing[d.Name] = true
	}

	uncovered := b.getUncoveredPackages(dir, existing)
	if len(uncovered) > 0 {
		deps = append(deps, b.lookupLatestVersions(dir, uncovered)...)
	}

	return deps
}

// fallbackPackageJSON is used when bun outdated fails entirely.
// Looks up latest versions from the npm registry for all packages.
func (b *BunScanner) fallbackPackageJSON(dir string) []Dependency {
	uncovered := b.getUncoveredPackages(dir, nil)
	if len(uncovered) == 0 {
		return nil
	}
	return b.lookupLatestVersions(dir, uncovered)
}

type uncoveredPkg struct {
	name string
	ver  string
	dev  bool
}

func (b *BunScanner) getUncoveredPackages(dir string, existing map[string]bool) []uncoveredPkg {
	pkgPath := filepath.Join(dir, "package.json")
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return nil
	}

	var uncovered []uncoveredPkg
	for name, ver := range pkg.Dependencies {
		if existing == nil || !existing[name] {
			uncovered = append(uncovered, uncoveredPkg{name, ver, false})
		}
	}
	for name, ver := range pkg.DevDependencies {
		if existing == nil || !existing[name] {
			uncovered = append(uncovered, uncoveredPkg{name, ver, true})
		}
	}
	return uncovered
}

func (b *BunScanner) lookupLatestVersions(dir string, uncovered []uncoveredPkg) []Dependency {
	// Use npm view to get latest versions (bun uses the npm registry)
	viewCmd := "npm"
	if !commandExists("npm") {
		viewCmd = "bun"
	}

	type result struct {
		latest string
	}
	results := make([]result, len(uncovered))
	var wg sync.WaitGroup
	for i, pkg := range uncovered {
		wg.Add(1)
		go func(idx int, pkgName string) {
			defer wg.Done()
			var latest string
			var err error
			if viewCmd == "npm" {
				latest, err = runCommand(dir, "npm", "view", pkgName, "version")
			} else {
				latest, err = runCommand(dir, "bun", "pm", "info", pkgName, "--json")
				// bun pm info returns JSON; extract version if needed
				if err == nil && latest != "" {
					var info struct {
						Version string `json:"version"`
					}
					if json.Unmarshal([]byte(latest), &info) == nil && info.Version != "" {
						latest = info.Version
					}
				}
			}
			if err == nil && latest != "" {
				results[idx] = result{latest: latest}
			}
		}(i, pkg.name)
	}
	wg.Wait()

	var deps []Dependency
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
			Ecosystem: EcosystemBun,
			Outdated:  current != latest,
			Indirect:  pkg.dev,
		})
	}
	return deps
}

func (b *BunScanner) Update(dir string, dep Dependency) error {
	_, err := runCommand(dir, "bun", "add", dep.Name+"@"+dep.Latest)
	return err
}

// splitLines splits a string into lines.
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// parseTableRow extracts cell values from a box-drawing table row like "│ foo │ bar │".
func parseTableRow(line string) []string {
	var fields []string
	inField := false
	fieldStart := 0

	runes := []rune(line)
	for i, r := range runes {
		if r == '│' || r == '|' {
			if inField {
				field := trimString(string(runes[fieldStart:i]))
				fields = append(fields, field)
			}
			inField = true
			fieldStart = i + 1
		}
	}
	return fields
}

func trimString(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// isSeparatorRow returns true for lines like "|------|------|" or "├──────┼──────┤".
func isSeparatorRow(line string) bool {
	for _, r := range line {
		switch r {
		case '|', '-', '─', '┌', '┐', '└', '┘', '├', '┤', '┬', '┴', '┼', '+', ' ', '\t', '\r':
			continue
		default:
			return false
		}
	}
	return true
}
