<p align="center">

```
    _    ___ _     _       _   
   / \  |_ _| |   (_)_ __ | |_ 
  / _ \  | || |   | | '_ \| __|
 / ___ \ | || |___| | | | | |_ 
/_/   \_\___|_____|_|_| |_|\__|
```

**The missing safety net for AI-generated code.**

[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev)
[![CI](https://github.com/SidharthSasikumar/ailint/actions/workflows/ci.yml/badge.svg)](https://github.com/SidharthSasikumar/ailint/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/SidharthSasikumar/ailint)](https://goreportcard.com/report/github.com/SidharthSasikumar/ailint)

</p>

---

Every developer is generating code with AI tools (Copilot, Claude Code, Cursor, Codex). But AI introduces a **specific class of bugs** that traditional linters miss completely: hallucinated packages, deprecated APIs from stale training data, casually generated secrets, and more.

**AILint catches them all.** Single binary. Zero runtime dependencies. Sub-100ms per scan.

## The 7 Deadly Sins of AI Code

| # | Rule | ID | What It Catches |
|---|------|----|-----------------|
| 1 | **Phantom Imports** | `phantom-imports` | Packages/modules that don't exist — AI hallucinated them |
| 2 | **Zombie APIs** | `zombie-apis` | Method calls that don't exist in the version being used |
| 3 | **Secret Leaks** | `secret-leaks` | Hardcoded credentials AI generates as "examples" |
| 4 | **Convention Drift** | `convention-drift` | Code patterns that deviate from your codebase conventions |
| 5 | **Stale Patterns** | `stale-patterns` | Deprecated APIs from AI's outdated training data |
| 6 | **License Risk** | `license-risk` | Generated code matching copyleft-licensed sources |
| 7 | **Complexity Bombs** | `complexity-bombs` | Over-engineered patterns where simple ones exist |

**MVP ships with:** Phantom Imports, Secret Leaks, and Stale Patterns.

## Quick Start

### Install

```bash
# Go install
go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest

# Homebrew (macOS/Linux)
brew install SidharthSasikumar/tap/ailint

# Download binary
curl -sSfL https://github.com/SidharthSasikumar/ailint/releases/latest/download/ailint_$(uname -s)_$(uname -m).tar.gz | tar xz
```

### Run

```bash
# Scan current directory
ailint

# Scan specific path
ailint ./src

# JSON output for CI
ailint -f json .

# SARIF output for GitHub Code Scanning
ailint -f sarif . > results.sarif
```

### Example Output

```
    _    ___ _     _       _   
   / \  |_ _| |   (_)_ __ | |_ 
  / _ \  | || |   | | '_ \| __|
 / ___ \ | || |___| | | | | |_ 
/_/   \_\___|_____|_|_| |_|\__|

  v0.1.0

internal/handler.go
  3:1   ✗ error    Package "github.com/golang/utils" not found in standard library or project dependencies  phantom-imports
         💡 Verify this package exists. AI tools sometimes hallucinate package names that look plausible but don't exist.
  14:1  ✗ error    AWS Access Key ID detected  secret-leaks
         💡 Use environment variables or a secrets manager instead of hardcoded values.

pkg/worker.go
  7:1   ⚠ warning  ioutil.ReadAll is deprecated since Go 1.16  stale-patterns
         💡 Use io.ReadAll instead

  2 error(s) · 1 warning(s) · 12 files scanned · 3 rules applied

  Trust Score: 62/100 (C)
```

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                         CLI (cmd/ailint)                      │
│  Flags → Config loader → Engine orchestration → Reporter     │
└──────────┬───────────────────┬──────────────────┬────────────┘
           │                   │                  │
    ┌──────▼──────┐    ┌───────▼───────┐   ┌──────▼──────┐
    │   Scanner   │    │    Engine     │   │  Reporter   │
    │             │    │              │   │             │
    │ • Walk dirs │    │ • Worker pool│   │ • Terminal  │
    │ • Git diff  │    │ • Fan-out    │   │ • JSON     │
    │ • Filter    │    │ • Aggregate  │   │ • SARIF    │
    └──────┬──────┘    └───────┬───────┘   └─────────────┘
           │                   │
           │            ┌──────▼──────┐
           │            │    Rules    │
           └───────────►│             │
                        │ ┌─────────┐ │
        ┌──────────┐    │ │Phantom  │ │
        │ Parsers  │◄───┤ │Imports  │ │
        │          │    │ ├─────────┤ │
        │ • Go     │    │ │Secret   │ │
        │ • Python │    │ │Leaks    │ │
        │ • JS/TS  │    │ ├─────────┤ │
        └──────────┘    │ │Stale    │ │
                        │ │Patterns │ │
                        │ └─────────┘ │
                        └─────────────┘
```

### Component Flow

1. **Scanner** discovers files (walk directory or parse git diff)
2. **Engine** fans files out to a worker pool
3. Each **Rule** analyzes files using language-specific **Parsers**
4. Findings are aggregated, scored, and passed to the **Reporter**
5. **Trust Score** (0–100) is calculated from finding severity

### Core Interfaces

```go
// Rule — implement this to add custom checks
type Rule interface {
    ID() string
    Name() string
    Description() string
    DefaultSeverity() types.Severity
    Languages() []string  // empty = all languages
    Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error)
}

// Parser — implement this to add language support
type Parser interface {
    Language() string
    Extensions() []string
    ParseImports(content []byte) []types.Import
}

// Reporter — implement this to add output formats
type Reporter interface {
    Format() string
    Report(result *types.Result, w io.Writer) error
}
```

## Configuration

Create `.ailint.yaml` in your project root:

```yaml
version: 1

rules:
  phantom-imports:
    enabled: true
    severity: error
    languages: [go, python, javascript]

  secret-leaks:
    enabled: true
    severity: error
    entropy_threshold: 4.5    # Shannon entropy threshold for secret detection

  stale-patterns:
    enabled: true
    severity: warning

  zombie-apis:
    enabled: false
    severity: warning

  convention-drift:
    enabled: false
    severity: info

  license-risk:
    enabled: false
    severity: warning

  complexity-bombs:
    enabled: false
    severity: info

output:
  format: terminal    # terminal, json, sarif
  color: true
  verbose: false

scan:
  paths: [.]
  exclude:
    - vendor/
    - node_modules/
    - .git/
    - dist/
    - build/
  diff_only: false    # Only scan files in git diff

trust_score:
  enabled: true
  thresholds:
    pass: 80          # Exit 0 if score >= 80
    warn: 60          # Warning if score < 60
```

### CLI Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--config` | `-c` | `.ailint.yaml` | Config file path |
| `--format` | `-f` | `terminal` | Output format: `terminal`, `json`, `sarif` |
| `--no-color` | | `false` | Disable colored output |
| `--workers` | `-j` | `NumCPU` | Parallel worker count |
| `--version` | `-v` | | Show version |

## CI Integration

### GitHub Actions

```yaml
# .github/workflows/ailint.yml
name: AILint
on: [push, pull_request]

jobs:
  ailint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Install AILint
        run: go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest

      - name: Run AILint
        run: ailint -f terminal .
```

### GitHub Actions with SARIF (Code Scanning)

```yaml
jobs:
  ailint:
    runs-on: ubuntu-latest
    permissions:
      security-events: write
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - run: go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest

      - name: Run AILint (SARIF)
        run: ailint -f sarif . > results.sarif
        continue-on-error: true

      - name: Upload SARIF
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: results.sarif
```

### Git Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit
ailint -f terminal --no-color .
```

Or with [pre-commit](https://pre-commit.com/):

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/SidharthSasikumar/ailint
    rev: v0.1.0
    hooks:
      - id: ailint
        entry: ailint
        language: golang
        types: [file]
        pass_filenames: false
```

### GitLab CI

```yaml
ailint:
  image: golang:1.24
  script:
    - go install github.com/SidharthSasikumar/ailint/cmd/ailint@latest
    - ailint -f json . > ailint-report.json
  artifacts:
    reports:
      codequality: ailint-report.json
```

## Trust Score

Every scan produces a **Trust Score** (0–100) that rates the confidence level of the analyzed code:

| Score | Grade | Meaning |
|-------|-------|---------|
| 90–100 | A | Code looks clean — minimal AI artifacts |
| 80–89 | B | Minor issues — review suggested findings |
| 60–79 | C | Moderate issues — several AI patterns detected |
| 40–59 | D | Significant issues — careful review required |
| 0–39 | F | Critical issues — likely unreviewed AI output |

**Scoring formula:**
- Each `error` finding: **-15 points**
- Each `warning` finding: **-8 points**  
- Each `info` finding: **-3 points**

## How Each Rule Works

### Phantom Imports

Parses import statements using language-specific parsers, then validates them against:
- **Go:** Standard library + `go.mod` dependencies + module path
- **Python:** Standard library + `requirements.txt` / `pyproject.toml`
- **JavaScript:** Node.js built-ins + `package.json` dependencies

AI tools frequently hallucinate package names that look plausible:
```go
import "github.com/golang/utils"      // ← doesn't exist
import "github.com/google/go-cloud"    // ← it's gocloud.dev
```

### Secret Leaks

Dual detection strategy:
1. **Pattern matching:** Regexes for AWS keys, GitHub tokens, Stripe keys, private key headers, and more
2. **Entropy analysis:** Shannon entropy calculation flags high-entropy strings in secret-like variable assignments

AI tools often generate realistic-looking credentials as "examples" that developers forget to replace.

### Stale Patterns

Embedded database of deprecated APIs across Go, Python, and JavaScript. AI models trained on older code frequently generate:
- `ioutil.ReadAll` instead of `io.ReadAll` (Go 1.16+)
- `typing.Dict` instead of `dict` (Python 3.9+)
- `new Buffer()` instead of `Buffer.from()` (Node.js 6+)
- `require('request')` — abandoned since 2020

## Project Structure

```
ailint/
├── cmd/
│   └── ailint/
│       └── main.go              # CLI entrypoint
├── internal/
│   ├── config/
│   │   └── config.go            # YAML config loader
│   ├── engine/
│   │   └── engine.go            # Analysis orchestration + worker pool
│   ├── parser/
│   │   ├── parser.go            # Parser interface + registry
│   │   ├── golang.go            # Go import parser + stdlib list
│   │   ├── python.go            # Python import parser + stdlib list
│   │   └── javascript.go        # JS/TS import parser + Node builtins
│   ├── reporter/
│   │   ├── reporter.go          # Reporter interface
│   │   ├── terminal.go          # Pretty terminal output with colors
│   │   ├── json.go              # JSON output
│   │   └── sarif.go             # SARIF 2.1.0 for GitHub Code Scanning
│   ├── rule/
│   │   ├── rule.go              # Rule interface
│   │   ├── phantom_imports.go   # Hallucinated package detection
│   │   ├── secret_leaks.go      # Credential leak detection
│   │   └── stale_patterns.go    # Deprecated API detection
│   └── scanner/
│       └── scanner.go           # File discovery + filtering
├── pkg/
│   └── types/
│       └── types.go             # Shared types (Finding, Severity, etc.)
├── .ailint.yaml                 # Example configuration
├── .github/workflows/ci.yml     # CI pipeline
├── .goreleaser.yml              # Release automation
├── action.yml                   # GitHub Action definition
├── go.mod
├── Makefile
├── LICENSE
└── README.md
```

## Roadmap

### v0.1.0 — MVP
- [x] Phantom Imports (Go, Python, JavaScript)
- [x] Secret Leaks (pattern + entropy detection)
- [x] Stale Patterns (deprecated API database)
- [x] Terminal, JSON, SARIF output
- [x] Trust Score
- [x] GitHub Action

### v0.2.0
- [ ] Zombie APIs — validate method calls against API docs
- [ ] Git diff mode — only scan changed files
- [ ] VS Code extension
- [ ] Pre-commit hook package

### v0.3.0
- [ ] Convention Drift — learn codebase patterns via OpenAI
- [ ] Complexity Bombs — detect over-engineering
- [ ] License Risk — copyleft source matching
- [ ] Custom rule SDK

### v1.0.0
- [ ] Language Server Protocol (LSP) support
- [ ] IDE plugins (VS Code, JetBrains)
- [ ] Rule marketplace
- [ ] Team configuration sharing

## Building from Source

```bash
git clone https://github.com/SidharthSasikumar/ailint.git
cd ailint
make build
./bin/ailint --version
```

## Contributing

Contributions welcome. Please open an issue first to discuss what you'd like to change.

```bash
# Run tests
make test

# Run linter
make lint

# Build
make build
```

## License

[MIT](LICENSE)

---

<p align="center">
Built by <a href="https://github.com/SidharthSasikumar">Sidharth Sasikumar</a> — because AI writes bugs too.
</p>
