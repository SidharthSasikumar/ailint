package reporter

import (
	"encoding/json"
	"io"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// JSONReporter writes results as JSON.
type JSONReporter struct{}

func (r *JSONReporter) Format() string { return "json" }

func (r *JSONReporter) Report(result *types.Result, w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
