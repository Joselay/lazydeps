# lazydeps

A terminal UI for scanning, viewing, and updating dependencies across multiple ecosystems.

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/Joselay/lazydeps)
![Release](https://img.shields.io/github/v/release/Joselay/lazydeps)

## Supported Ecosystems

| Ecosystem | Detect | Outdated | Vulnerabilities | Update |
|-----------|--------|----------|-----------------|--------|
| Go        | `go.mod` | `go list -m -u` | `govulncheck` | `go get` |
| npm       | `package.json` | `npm outdated` | `npm audit` | `npm install` |
| Bun       | `bun.lock` | `bun outdated` | - | `bun add` |
| pip       | `requirements.txt` / `pyproject.toml` | `pip3 list --outdated` | - | `pip3 install --upgrade` |
| Cargo     | `Cargo.toml` | `cargo outdated` | `cargo audit` | `cargo update` |

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap Joselay/tap
brew install lazydeps
```

### Download binary

Grab the latest binary from [Releases](https://github.com/Joselay/lazydeps/releases) and add it to your `PATH`.

### From source (requires Go 1.26+)

```bash
go install github.com/Joselay/lazydeps@latest
```

## Usage

```bash
# Scan the current directory
lazydeps

# Scan a specific project
lazydeps /path/to/project
```

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` or `↑` / `↓` | Navigate |
| `Tab` / `Shift+Tab` | Switch tabs (All / Outdated / Vulnerable) |
| `Enter` | View dependency details |
| `u` | Update selected dependency |
| `U` | Update all outdated dependencies |
| `/` | Filter dependencies |
| `r` | Refresh scan |
| `v` | Toggle vulnerable only |
| `?` | Toggle help |
| `q` | Quit |

## License

[MIT](LICENSE)
