# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[0.1.0]: https://github.com/Joselay/lazydeps/releases/tag/v0.1.0
