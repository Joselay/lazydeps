package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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
	Name           string `json:"name"`
	Current        string `json:"current"`
	Update         string `json:"update"`
	Latest         string `json:"latest"`
}

func (b *BunScanner) Scan(dir string) ([]Dependency, error) {
	if !commandExists("bun") {
		return nil, nil
	}

	var deps []Dependency

	// bun outdated outputs a table by default; no --json yet,
	// so we parse package.json and use bun pm ls for installed versions
	deps = b.scanOutdated(dir)

	// Fall back to package.json parsing if nothing returned
	if len(deps) == 0 {
		deps = b.parsePackageJSON(dir)
	}

	return deps, nil
}

func (b *BunScanner) scanOutdated(dir string) []Dependency {
	// bun outdated (available since bun 1.1.28) outputs text table
	out, err := runCommand(dir, "bun", "outdated")
	if err != nil || out == "" {
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

	// Also add up-to-date deps from package.json
	existing := make(map[string]bool)
	for _, d := range deps {
		existing[d.Name] = true
	}
	for _, d := range b.parsePackageJSON(dir) {
		if !existing[d.Name] {
			deps = append(deps, d)
		}
	}

	return deps
}

func (b *BunScanner) parsePackageJSON(dir string) []Dependency {
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

	var deps []Dependency
	for name, ver := range pkg.Dependencies {
		deps = append(deps, Dependency{
			Name:      name,
			Current:   ver,
			Latest:    ver,
			Ecosystem: EcosystemBun,
		})
	}
	for name, ver := range pkg.DevDependencies {
		deps = append(deps, Dependency{
			Name:      name,
			Current:   ver,
			Latest:    ver,
			Ecosystem: EcosystemBun,
			Indirect:  true,
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
