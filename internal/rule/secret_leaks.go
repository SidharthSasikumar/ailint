package rule

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// SecretLeaks catches hardcoded credentials and secrets that AI tools
// generate as "examples" — using both pattern matching and entropy analysis.
type SecretLeaks struct {
	entropyThreshold float64
	patterns         []*secretPattern
}

type secretPattern struct {
	name    string
	pattern *regexp.Regexp
	message string
}

// NewSecretLeaks creates a SecretLeaks rule with the given entropy threshold.
func NewSecretLeaks(entropyThreshold float64) *SecretLeaks {
	if entropyThreshold <= 0 {
		entropyThreshold = 4.5
	}
	return &SecretLeaks{
		entropyThreshold: entropyThreshold,
		patterns:         defaultSecretPatterns(),
	}
}

func (r *SecretLeaks) ID() string                      { return "secret-leaks" }
func (r *SecretLeaks) Name() string                    { return "Secret Leaks" }
func (r *SecretLeaks) DefaultSeverity() types.Severity { return types.SeverityError }
func (r *SecretLeaks) Languages() []string             { return nil } // All languages
func (r *SecretLeaks) Description() string {
	return "Catches hardcoded credentials and secrets that AI tools generate as examples"
}

func (r *SecretLeaks) Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error) {
	var findings []types.Finding

	for i, line := range file.Lines {
		lineNum := i + 1

		// Skip comment-only lines (basic heuristic)
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") && strings.Contains(strings.ToLower(trimmed), "example") {
			continue
		}

		// Pattern-based detection
		for _, p := range r.patterns {
			if p.pattern.MatchString(line) {
				findings = append(findings, types.Finding{
					RuleID:     r.ID(),
					RuleName:   r.Name(),
					Severity:   r.DefaultSeverity(),
					File:       file.Path,
					Line:       lineNum,
					Column:     1,
					Message:    p.message,
					Suggestion: "Use environment variables or a secrets manager instead of hardcoded values.",
				})
			}
		}

		// Entropy-based detection for secret-looking assignments
		if finding := r.checkEntropy(file.Path, line, lineNum); finding != nil {
			findings = append(findings, *finding)
		}
	}

	return r.deduplicate(findings), nil
}

// checkEntropy looks for high-entropy strings in secret-like variable assignments.
func (r *SecretLeaks) checkEntropy(filePath, line string, lineNum int) *types.Finding {
	m := secretAssignment.FindStringSubmatch(line)
	if m == nil {
		return nil
	}

	value := m[1]
	if len(value) < 8 {
		return nil
	}

	// Skip common placeholder values
	lower := strings.ToLower(value)
	for _, skip := range []string{"changeme", "placeholder", "your-", "xxx", "todo", "fixme", "replace"} {
		if strings.Contains(lower, skip) {
			return nil
		}
	}

	entropy := shannonEntropy(value)
	if entropy >= r.entropyThreshold {
		return &types.Finding{
			RuleID:   r.ID(),
			RuleName: r.Name(),
			Severity: types.SeverityError,
			File:     filePath,
			Line:     lineNum,
			Column:   strings.Index(line, value) + 1,
			Message: fmt.Sprintf(
				"Potential hardcoded secret detected (entropy: %.1f, threshold: %.1f)",
				entropy, r.entropyThreshold),
			Suggestion: "Use environment variables or a secrets manager. AI tools often generate realistic-looking credentials that get committed.",
		}
	}
	return nil
}

// deduplicate removes findings that flag the same line for both pattern and entropy.
func (r *SecretLeaks) deduplicate(findings []types.Finding) []types.Finding {
	seen := map[string]bool{}
	var result []types.Finding
	for _, f := range findings {
		key := fmt.Sprintf("%s:%d", f.File, f.Line)
		if !seen[key] {
			seen[key] = true
			result = append(result, f)
		}
	}
	return result
}

var secretAssignment = regexp.MustCompile(
	`(?i)(?:password|passwd|pwd|secret|token|api[_-]?key|apikey|auth[_-]?token|credential|private[_-]?key|access[_-]?key)` +
		`\s*[:=]\s*["']([^"']{8,})["']`)

// shannonEntropy calculates the Shannon entropy of a string in bits.
func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, c := range s {
		freq[c]++
	}
	length := float64(len([]rune(s)))
	var entropy float64
	for _, count := range freq {
		p := count / length
		if p > 0 {
			entropy -= p * math.Log2(p)
		}
	}
	return entropy
}

func defaultSecretPatterns() []*secretPattern {
	return []*secretPattern{
		{
			name:    "AWS Access Key",
			pattern: regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
			message: "AWS Access Key ID detected",
		},
		{
			name:    "AWS Secret Key",
			pattern: regexp.MustCompile(`(?i)aws.?secret.?access.?key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`),
			message: "AWS Secret Access Key detected",
		},
		{
			name:    "GitHub Token",
			pattern: regexp.MustCompile(`gh[pousr]_[A-Za-z0-9_]{36,}`),
			message: "GitHub personal access token detected",
		},
		{
			name:    "Generic API Key Assignment",
			pattern: regexp.MustCompile(`(?i)(?:api[_-]?key|apikey)\s*[:=]\s*["']([A-Za-z0-9]{32,})["']`),
			message: "Potential API key detected in assignment",
		},
		{
			name:    "Private Key Header",
			pattern: regexp.MustCompile(`-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`),
			message: "Private key detected in source file",
		},
		{
			name:    "Slack Token",
			pattern: regexp.MustCompile(`xox[baprs]-[0-9]{10,13}-[0-9]{10,13}[a-zA-Z0-9-]*`),
			message: "Slack token detected",
		},
		{
			name:    "Google API Key",
			pattern: regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),
			message: "Google API key detected",
		},
		{
			name:    "Stripe Key",
			pattern: regexp.MustCompile(`(?:sk|pk)_(?:test|live)_[A-Za-z0-9]{20,}`),
			message: "Stripe API key detected",
		},
		{
			name:    "Heroku API Key",
			pattern: regexp.MustCompile(`(?i)heroku.?api.?key\s*[:=]\s*["']?[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}["']?`),
			message: "Heroku API key detected",
		},
		{
			name:    "Generic High-Entropy Secret",
			pattern: regexp.MustCompile(`(?i)(?:secret|password|passwd|pwd)\s*[:=]\s*["']([^"']{16,})["']`),
			message: "Potential hardcoded secret in assignment",
		},
	}
}
