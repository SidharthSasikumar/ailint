package reporter

import (
	"io"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Reporter formats analysis results for output.
type Reporter interface {
	// Format returns the output format name.
	Format() string

	// Report writes the formatted results to the writer.
	Report(result *types.Result, w io.Writer) error
}
