package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/SidharthSasikumar/ailint/internal/parser"
	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// PhantomImports detects imports of packages that don't exist —
// a common class of bugs where AI tools hallucinate package names.
type PhantomImports struct {
	goModDeps map[string]bool
	goModPath string
	pyDeps    map[string]bool
	jsDeps    map[string]bool
	hasGoMod  bool
	hasPyDeps bool
	hasJSDeps bool
}

// NewPhantomImports creates a PhantomImports rule, loading dependency
// manifests from the given root directory.
func NewPhantomImports(rootDir string) *PhantomImports {
	r := &PhantomImports{
		goModDeps: make(map[string]bool),
		pyDeps:    make(map[string]bool),
		jsDeps:    make(map[string]bool),
	}
	r.loadGoMod(rootDir)
	r.loadPythonDeps(rootDir)
	r.loadJSDeps(rootDir)
	return r
}

func (r *PhantomImports) ID() string                      { return "phantom-imports" }
func (r *PhantomImports) Name() string                    { return "Phantom Imports" }
func (r *PhantomImports) DefaultSeverity() types.Severity { return types.SeverityError }
func (r *PhantomImports) Languages() []string             { return []string{"go", "python", "javascript"} }
func (r *PhantomImports) Description() string {
	return "Detects imports of packages/modules that don't exist (AI hallucinated them)"
}

func (r *PhantomImports) Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error) {
	p := parser.ForLanguage(file.Language)
	if p == nil {
		return nil, nil
	}

	imports := p.ParseImports(file.Content)
	var findings []types.Finding

	for _, imp := range imports {
		if r.isPhantom(file.Language, imp) {
			findings = append(findings, types.Finding{
				RuleID:   r.ID(),
				RuleName: r.Name(),
				Severity: r.DefaultSeverity(),
				File:     file.Path,
				Line:     imp.Line,
				Column:   1,
				Message: fmt.Sprintf("Package %q not found in standard library or project dependencies",
					imp.Path),
				Suggestion: "Verify this package exists. AI tools sometimes hallucinate package names that look plausible but don't exist.",
			})
		}
	}

	return findings, nil
}

func (r *PhantomImports) isPhantom(lang string, imp types.Import) bool {
	switch lang {
	case "go":
		return r.isPhantomGo(imp)
	case "python":
		return r.isPhantomPython(imp)
	case "javascript":
		return r.isPhantomJS(imp)
	}
	return false
}

func (r *PhantomImports) isPhantomGo(imp types.Import) bool {
	// Check Go standard library
	if parser.GoStdlib[imp.Path] {
		return false
	}
	// Check stdlib sub-packages
	for pkg := range parser.GoStdlib {
		if strings.HasPrefix(imp.Path, pkg+"/") {
			return false
		}
	}

	// Without go.mod, we can only flag single-segment imports that aren't stdlib.
	// Multi-segment imports (github.com/...) can't be verified without a manifest.
	if !r.hasGoMod {
		return !strings.Contains(imp.Path, "/") && !strings.Contains(imp.Path, ".")
	}

	// Check go.mod dependencies (includes module path for internal imports)
	for dep := range r.goModDeps {
		if imp.Path == dep || strings.HasPrefix(imp.Path, dep+"/") {
			return false
		}
	}

	return true
}

func (r *PhantomImports) isPhantomPython(imp types.Import) bool {
	// Relative imports are always fine
	if strings.HasPrefix(imp.Path, ".") {
		return false
	}
	// Standard library
	if parser.PythonStdlib[imp.Name] {
		return false
	}
	// Without a manifest, we can't verify third-party imports
	if !r.hasPyDeps {
		return false
	}
	// Check declared dependencies (case-insensitive, normalize hyphens)
	normalized := strings.ToLower(strings.ReplaceAll(imp.Name, "-", "_"))
	return !r.pyDeps[normalized]
}

func (r *PhantomImports) isPhantomJS(imp types.Import) bool {
	// Node.js builtins
	if parser.NodeBuiltins[imp.Name] || parser.NodeBuiltins[imp.Path] {
		return false
	}
	// Without package.json, we can't verify
	if !r.hasJSDeps {
		return false
	}
	return !r.jsDeps[imp.Name]
}

// loadGoMod parses go.mod to extract the module path and dependencies.
func (r *PhantomImports) loadGoMod(root string) {
	data, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return
	}
	r.hasGoMod = true

	lines := strings.Split(string(data), "\n")
	inRequire := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract module path
		if strings.HasPrefix(trimmed, "module ") {
			r.goModPath = strings.TrimSpace(strings.TrimPrefix(trimmed, "module "))
			r.goModDeps[r.goModPath] = true
		}

		// Parse require blocks
		if trimmed == "require (" {
			inRequire = true
			continue
		}
		if inRequire && trimmed == ")" {
			inRequire = false
			continue
		}
		if inRequire {
			parts := strings.Fields(trimmed)
			if len(parts) >= 1 && !strings.HasPrefix(parts[0], "//") {
				r.goModDeps[parts[0]] = true
			}
		}

		// Single-line require
		if strings.HasPrefix(trimmed, "require ") && !strings.Contains(trimmed, "(") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				r.goModDeps[parts[1]] = true
			}
		}
	}
}

// loadPythonDeps reads requirements.txt and pyproject.toml.
func (r *PhantomImports) loadPythonDeps(root string) {
	// Try requirements.txt
	if data, err := os.ReadFile(filepath.Join(root, "requirements.txt")); err == nil {
		r.hasPyDeps = true
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
				continue
			}
			// Strip version specifiers and extras
			for _, sep := range []string{">=", "<=", "==", "!=", "~=", ">", "<", "["} {
				if idx := strings.Index(line, sep); idx > 0 {
					line = line[:idx]
				}
			}
			name := strings.TrimSpace(strings.ReplaceAll(line, "-", "_"))
			r.pyDeps[strings.ToLower(name)] = true
		}
	}

	// Try pyproject.toml dependencies array
	if data, err := os.ReadFile(filepath.Join(root, "pyproject.toml")); err == nil {
		r.hasPyDeps = true
		for _, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "\"") || strings.HasPrefix(trimmed, "'") {
				// Extract package name from lines like: "flask>=2.0"
				name := strings.Trim(trimmed, "\"',")
				for _, sep := range []string{">=", "<=", "==", "!=", "~=", ">", "<", "["} {
					if idx := strings.Index(name, sep); idx > 0 {
						name = name[:idx]
					}
				}
				name = strings.TrimSpace(strings.ReplaceAll(name, "-", "_"))
				if name != "" {
					r.pyDeps[strings.ToLower(name)] = true
				}
			}
		}
	}
}

// loadJSDeps reads package.json dependencies.
func (r *PhantomImports) loadJSDeps(root string) {
	data, err := os.ReadFile(filepath.Join(root, "package.json"))
	if err != nil {
		return
	}
	r.hasJSDeps = true

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if json.Unmarshal(data, &pkg) != nil {
		return
	}
	for name := range pkg.Dependencies {
		r.jsDeps[name] = true
	}
	for name := range pkg.DevDependencies {
		r.jsDeps[name] = true
	}
}
