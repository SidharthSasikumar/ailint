# ailint

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![CI](https://github.com/SidharthSasikumar/ailint/actions/workflows/ci.yml/badge.svg)](https://github.com/SidharthSasikumar/ailint/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/SidharthSasikumar/ailint)](https://goreportcard.com/report/github.com/SidharthSasikumar/ailint)

A static analysis tool that catches bugs specific to AI-generated code — hallucinated imports, leaked secrets, deprecated APIs — that traditional linters don't flag.

Single binary, zero runtime dependencies, supports Go/Python/JavaScript.

## Rules

| Rule | ID | Status |
|------|----|--------|
| Phantom Imports | `phantom-imports` | ✅ Detects imports of packages that don't exist |
| Secret Leaks | `secret-leaks` | ✅ Hardcoded credentials via pattern matching + Shannon entropy |
| Stale Patterns | `stale-patterns` | ✅ Deprecated APIs from outdated training data |
| Zombie APIs | `zombie-apis` | 🔜 Method calls that don't exist in the version used |
| Convention Drift | `convention-drift` | 🔜 Patterns deviating from codebase conventions |
| License Risk | `license-risk` | 🔜 Code matching copyleft-licensed sources |
| Complexity Bombs | `complexity-bombs` | 🔜 Over-engineered patterns where simple ones exist |

## Install

```bash
go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest
```

Or download a binary from [Releases](https://github.com/SidharthSasikumar/ailint/releases).

## Usage

```bash
ailint                      # scan current directory
ailint ./src                # scan specific path
ailint -f json .            # JSON output
ailint -f sarif . > r.sarif # SARIF for GitHub Code Scanning
ailint -j 4 .               # limit to 4 workers
```

### Output

```
internal/handler.go
  3:1   ✗ error    Package "github.com/golang/utils" not found in standard library
         or project dependencies  phantom-imports
         💡 Verify this package exists.
  14:1  ✗ error    AWS Access Key ID detected  secret-leaks
         💡 Use environment variables or a secrets manager instead.

pkg/worker.go
  7:1   ⚠ warning  ioutil.ReadAll is deprecated since Go 1.16  stale-patterns
         💡 Use io.ReadAll instead

  2 error(s) · 1 warning(s) · 12 files scanned · 3 rules applied

  Trust Score: 62/100 (C)
```

## Configuration

Drop a `.ailint.yaml` in your project root:

```yaml
version: 1

rules:
  phantom-imports:
    enabled: true
    severity: error

  secret-leaks:
    enabled: true
    severity: error
    entropy_threshold: 4.5

  stale-patterns:
    enabled: true
    severity: warning

output:
  format: terminal
  color: true

scan:
  exclude:
    - vendor/
    - node_modules/
    - .git/
    - dist/

trust_score:
  enabled: true
  thresholds:
    pass: 80
    warn: 60
```

Without a config file, all three MVP rules run with sensible defaults.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `.ailint.yaml` | Config file path |
| `--format` | `-f` | `terminal` | Output: `terminal`, `json`, `sarif` |
| `--no-color` | | `false` | Disable colored output |
| `--workers` | `-j` | `NumCPU` | Parallel worker count |
| `--version` | `-v` | | Show version |

## CI

### GitHub Actions

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: '1.24'
- run: go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest
- run: ailint .
```

For SARIF integration with GitHub Code Scanning:

```yaml
- run: ailint -f sarif . > results.sarif
  continue-on-error: true
- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: results.sarif
```

### Pre-commit

```bash
#!/bin/sh
# .git/hooks/pre-commit
ailint --no-color .
```

## Trust Score

Every scan produces a score from 0 to 100:

| Grade | Score | Meaning |
|-------|-------|---------|
| A | 90–100 | Clean |
| B | 80–89 | Minor issues |
| C | 60–79 | Needs review |
| D | 40–59 | Significant issues |
| F | 0–39 | Unreviewed AI output |

Penalties: **-15** per error, **-8** per warning, **-3** per info.

## How It Works

**Phantom Imports** — parses import statements and validates against the project's dependency manifest (`go.mod`, `requirements.txt`, `package.json`) and the language's standard library. Imports that resolve to nothing get flagged.

**Secret Leaks** — dual approach: regex patterns for known credential formats (AWS keys, GitHub tokens, private key headers, etc.) plus Shannon entropy analysis on values assigned to secret-looking variables.

**Stale Patterns** — embedded database of ~25 deprecated APIs across Go, Python, and JS with specific replacements (e.g., `ioutil.ReadAll` → `io.ReadAll`).

## Building

```bash
git clone https://github.com/SidharthSasikumar/ailint.git
cd ailint
make build       # → bin/ailint
make test        # run tests
make lint        # golangci-lint
```

## Contributing

Open an issue first to discuss the change. PRs welcome.

## License

[MIT](LICENSE)
