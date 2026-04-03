package reporter

import (
	"io"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Reporter writes analysis results to an output stream.
type Reporter interface {
	Format() string
	Report(result *types.Result, w io.Writer) error
}
