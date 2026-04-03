package rule

import (
	"context"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Rule defines the interface for all lint rules.
type Rule interface {
	// ID returns the unique identifier for this rule (e.g., "phantom-imports").
	ID() string

	// Name returns the human-readable rule name.
	Name() string

	// Description returns a short explanation of what the rule checks.
	Description() string

	// DefaultSeverity returns the default severity level.
	DefaultSeverity() types.Severity

	// Languages returns the languages this rule applies to.
	// An empty slice means the rule applies to all languages.
	Languages() []string

	// Check analyzes a file and returns any findings.
	Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error)
}
