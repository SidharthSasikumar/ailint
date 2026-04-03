package parser

import (
	"regexp"
	"strings"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// JavaScriptParser handles JS/TS source files.
type JavaScriptParser struct{}

func (p *JavaScriptParser) Language() string { return "javascript" }
func (p *JavaScriptParser) Extensions() []string {
	return []string{".js", ".jsx", ".ts", ".tsx", ".mjs", ".cjs"}
}

var jsImportFrom = regexp.MustCompile(`import\s+.*?\s+from\s+['"]([^'"]+)['"]`)
var jsImportPlain = regexp.MustCompile(`import\s+['"]([^'"]+)['"]`)
var jsRequire = regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`)

func (p *JavaScriptParser) ParseImports(content []byte) []types.Import {
	var imports []types.Import
	seen := map[string]bool{}

	for i, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)

		// Skip comments
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
			continue
		}

		for _, re := range []*regexp.Regexp{jsImportFrom, jsImportPlain, jsRequire} {
			if m := re.FindStringSubmatch(line); m != nil {
				pkg := m[1]

				// Skip relative imports
				if strings.HasPrefix(pkg, ".") || strings.HasPrefix(pkg, "/") {
					continue
				}

				// Extract package name (handle scoped packages)
				name := jsPackageName(pkg)
				if seen[name] {
					continue
				}
				seen[name] = true

				imports = append(imports, types.Import{
					Path: pkg,
					Name: name,
					Line: i + 1,
				})
			}
		}
	}
	return imports
}

// jsPackageName extracts the npm package name from a path.
func jsPackageName(path string) string {
	// Scoped packages: @scope/package/sub → @scope/package
	if strings.HasPrefix(path, "@") {
		parts := strings.SplitN(path, "/", 3)
		if len(parts) >= 2 {
			return parts[0] + "/" + parts[1]
		}
		return path
	}
	// Regular packages: package/sub → package
	return strings.SplitN(path, "/", 2)[0]
}

// NodeBuiltins lists Node.js built-in modules.
var NodeBuiltins = map[string]bool{
	"assert": true, "assert/strict": true, "async_hooks": true,
	"buffer": true, "child_process": true, "cluster": true,
	"console": true, "constants": true, "crypto": true,
	"dgram": true, "diagnostics_channel": true, "dns": true, "dns/promises": true,
	"domain": true, "events": true, "fs": true, "fs/promises": true,
	"http": true, "http2": true, "https": true,
	"inspector": true, "inspector/promises": true,
	"module": true, "net": true, "os": true, "path": true, "path/posix": true,
	"path/win32": true, "perf_hooks": true, "process": true,
	"punycode": true, "querystring": true,
	"readline": true, "readline/promises": true, "repl": true,
	"stream": true, "stream/consumers": true, "stream/promises": true,
	"stream/web": true, "string_decoder": true, "sys": true,
	"test": true, "timers": true, "timers/promises": true,
	"tls": true, "trace_events": true, "tty": true,
	"url": true, "util": true, "util/types": true,
	"v8": true, "vm": true, "wasi": true,
	"worker_threads": true, "zlib": true,
	// node: prefixed variants
	"node:assert": true, "node:buffer": true, "node:child_process": true,
	"node:cluster": true, "node:console": true, "node:crypto": true,
	"node:dgram": true, "node:diagnostics_channel": true, "node:dns": true,
	"node:events": true, "node:fs": true, "node:http": true, "node:http2": true,
	"node:https": true, "node:inspector": true, "node:module": true,
	"node:net": true, "node:os": true, "node:path": true, "node:perf_hooks": true,
	"node:process": true, "node:querystring": true, "node:readline": true,
	"node:repl": true, "node:stream": true, "node:string_decoder": true,
	"node:test": true, "node:timers": true, "node:tls": true, "node:trace_events": true,
	"node:tty": true, "node:url": true, "node:util": true, "node:v8": true,
	"node:vm": true, "node:wasi": true, "node:worker_threads": true, "node:zlib": true,
}
