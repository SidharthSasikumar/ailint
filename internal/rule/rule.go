package rule

import (
	"context"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Rule is the interface all lint rules implement.
type Rule interface {
	ID() string
	Name() string
	Description() string
	DefaultSeverity() types.Severity
	Languages() []string // empty = all languages
	Check(ctx context.Context, file *types.FileContext) ([]types.Finding, error)
}
