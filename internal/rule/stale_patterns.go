package rule

import (
	"context"
	"fmt"
	"regexp"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// StalePatterns flags usage of deprecated APIs.
type StalePatterns struct{}

// NewStalePatterns returns a configured StalePatterns rule.
func NewStalePatterns() *StalePatterns {
	return &StalePatterns{}
}

func (r *StalePatterns) ID() string                      { return "stale-patterns" }
func (r *StalePatterns) Name() string                    { return "Stale Patterns" }
func (r *StalePatterns) DefaultSeverity() types.Severity { return types.SeverityWarning }
func (r *StalePatterns) Languages() []string             { return nil } // All languages
func (r *StalePatterns) Description() string {
	return "Flags deprecated APIs that have standard replacements"
}

func (r *StalePatterns) Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error) {
	var findings []types.Finding

	for _, dep := range deprecatedAPIs {
		// Skip rules that don't apply to this language
		if dep.language != "" && dep.language != file.Language {
			continue
		}

		for i, line := range file.Lines {
			if dep.pattern.MatchString(line) {
				findings = append(findings, types.Finding{
					RuleID:   r.ID(),
					RuleName: r.Name(),
					Severity: r.DefaultSeverity(),
					File:     file.Path,
					Line:     i + 1,
					Column:   1,
					Message: fmt.Sprintf("%s is deprecated since %s",
						dep.name, dep.since),
					Suggestion: fmt.Sprintf("Use %s instead", dep.replacement),
				})
			}
		}
	}

	return findings, nil
}

type deprecatedAPI struct {
	language    string
	pattern     *regexp.Regexp
	name        string
	replacement string
	since       string
}

// deprecatedAPIs is the database of known deprecated APIs and their replacements.
var deprecatedAPIs = []deprecatedAPI{
	// ── Go ──────────────────────────────────────────────
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.ReadAll\b`),
		name:        "ioutil.ReadAll",
		replacement: "io.ReadAll",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.ReadFile\b`),
		name:        "ioutil.ReadFile",
		replacement: "os.ReadFile",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.WriteFile\b`),
		name:        "ioutil.WriteFile",
		replacement: "os.WriteFile",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.TempDir\b`),
		name:        "ioutil.TempDir",
		replacement: "os.MkdirTemp",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.TempFile\b`),
		name:        "ioutil.TempFile",
		replacement: "os.CreateTemp",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.ReadDir\b`),
		name:        "ioutil.ReadDir",
		replacement: "os.ReadDir",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.NopCloser\b`),
		name:        "ioutil.NopCloser",
		replacement: "io.NopCloser",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bioutil\.Discard\b`),
		name:        "ioutil.Discard",
		replacement: "io.Discard",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`"io/ioutil"`),
		name:        "io/ioutil package",
		replacement: "io and os packages",
		since:       "Go 1.16",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\binterface\s*\{\s*\}`),
		name:        "interface{}",
		replacement: "any (type alias)",
		since:       "Go 1.18",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bgolang\.org/x/exp/slices\b`),
		name:        "golang.org/x/exp/slices",
		replacement: "slices (stdlib)",
		since:       "Go 1.21",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bgolang\.org/x/exp/maps\b`),
		name:        "golang.org/x/exp/maps",
		replacement: "maps (stdlib)",
		since:       "Go 1.21",
	},
	{
		language:    "go",
		pattern:     regexp.MustCompile(`\bgolang\.org/x/exp/slog\b`),
		name:        "golang.org/x/exp/slog",
		replacement: "log/slog (stdlib)",
		since:       "Go 1.21",
	},

	// ── Python ──────────────────────────────────────────
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\bfrom\s+distutils\b`),
		name:        "distutils",
		replacement: "setuptools",
		since:       "Python 3.10 (removed in 3.12)",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\bimport\s+imp\b`),
		name:        "imp module",
		replacement: "importlib",
		since:       "Python 3.4",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\bimport\s+optparse\b`),
		name:        "optparse",
		replacement: "argparse",
		since:       "Python 3.2",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`@asyncio\.coroutine`),
		name:        "@asyncio.coroutine decorator",
		replacement: "async def",
		since:       "Python 3.8 (removed in 3.11)",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\bcollections\.(Mapping|MutableMapping|Sequence|MutableSequence)\b`),
		name:        "collections ABCs (direct access)",
		replacement: "collections.abc.*",
		since:       "Python 3.3 (removed in 3.10)",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\btyping\.(Dict|List|Tuple|Set|FrozenSet|Type)\b`),
		name:        "typing generic aliases",
		replacement: "built-in generics (dict, list, tuple, set)",
		since:       "Python 3.9",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\btyping\.Optional\b`),
		name:        "typing.Optional",
		replacement: "X | None syntax",
		since:       "Python 3.10",
	},
	{
		language:    "python",
		pattern:     regexp.MustCompile(`\btyping\.Union\b`),
		name:        "typing.Union",
		replacement: "X | Y syntax",
		since:       "Python 3.10",
	},

	// ── JavaScript/TypeScript ──────────────────────────
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`\.substr\(`),
		name:        "String.prototype.substr()",
		replacement: "String.prototype.slice()",
		since:       "MDN deprecated (Annex B)",
	},
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`\b__proto__\b`),
		name:        "__proto__",
		replacement: "Object.getPrototypeOf() / Object.setPrototypeOf()",
		since:       "ES6 (Annex B)",
	},
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`new\s+Buffer\s*\(`),
		name:        "new Buffer()",
		replacement: "Buffer.from() or Buffer.alloc()",
		since:       "Node.js 6 (security vulnerability)",
	},
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`require\s*\(\s*['"]request['"]\s*\)`),
		name:        "request module",
		replacement: "node-fetch, axios, or built-in fetch (Node 18+)",
		since:       "2020 (deprecated by maintainer)",
	},
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`require\s*\(\s*['"]moment['"]\s*\)`),
		name:        "moment.js",
		replacement: "date-fns, dayjs, or Temporal API",
		since:       "2020 (maintenance mode)",
	},
	{
		language:    "javascript",
		pattern:     regexp.MustCompile(`(?:from|require\s*\(\s*)['"]querystring['"]\)?`),
		name:        "querystring module",
		replacement: "URLSearchParams (Web API)",
		since:       "Node.js 14",
	},
}
