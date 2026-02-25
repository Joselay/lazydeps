package scanner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Scanner detects and scans dependencies for a given ecosystem.
type Scanner interface {
	Detect(dir string) bool
	Scan(dir string) ([]Dependency, error)
	Update(dir string, dep Dependency) error
	Name() Ecosystem
}

// ScanAll detects all ecosystems present in dir and scans them.
func ScanAll(dir string) ([]Dependency, []Ecosystem, error) {
	scanners := []Scanner{
		&GoScanner{},
		&BunScanner{},
		&NpmScanner{},
		&PipScanner{},
		&CargoScanner{},
	}

	var allDeps []Dependency
	var detected []Ecosystem

	for _, s := range scanners {
		if s.Detect(dir) {
			detected = append(detected, s.Name())
			deps, err := s.Scan(dir)
			if err != nil {
				continue
			}
			allDeps = append(allDeps, deps...)
		}
	}

	return allDeps, detected, nil
}

// GetUpdater returns the scanner for a given ecosystem.
func GetUpdater(eco Ecosystem) Scanner {
	switch eco {
	case EcosystemGo:
		return &GoScanner{}
	case EcosystemNpm:
		return &NpmScanner{}
	case EcosystemPip:
		return &PipScanner{}
	case EcosystemCargo:
		return &CargoScanner{}
	case EcosystemBun:
		return &BunScanner{}
	}
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runCommand(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("%s: %w\n%s", name, err, stderr.String())
	}
	return strings.TrimSpace(stdout.String()), nil
}
