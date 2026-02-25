# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-02-25

### Fixed

- npm scanner: stderr warnings corrupting JSON output from `npm outdated`
- npm scanner: devDependencies not detected as outdated when `node_modules` is absent
- Bun scanner: discarding outdated data due to non-zero exit code from `bun outdated`
- All scanners: version range prefixes (`^`, `~`, `>=`) no longer shown as current version

### Improved

- npm/Bun: parallel `npm view` lookups for packages not covered by outdated commands
- pip: parallel `pip3 index versions` lookups for packages only in requirements.txt
- Cargo: parallel `cargo search` lookups when `cargo-outdated` is not installed
- Separated stdout/stderr in command execution to prevent output corruption across all scanners

## [0.1.0] - 2026-02-25

### Added

- Interactive TUI for scanning and updating dependencies
- Support for 5 ecosystems: Go, npm, Bun, pip, and Cargo
- Auto-detection of project ecosystems
- Outdated dependency detection with current vs latest version comparison
- Vulnerability scanning via `govulncheck`, `npm audit`, and `cargo audit`
- Update individual or all outdated dependencies from the TUI
- Tab views: All, Outdated, Vulnerable
- Live text filtering with `/`
- Vim-style navigation (`j`/`k`/`h`/`l`)
- Detail view for individual dependencies
- Cross-platform release binaries via GoReleaser

[0.2.0]: https://github.com/Joselay/lazydeps/releases/tag/v0.2.0
[0.1.0]: https://github.com/Joselay/lazydeps/releases/tag/v0.1.0
