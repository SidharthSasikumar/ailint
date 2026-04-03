package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// TerminalReporter outputs results as pretty, human-readable terminal output.
type TerminalReporter struct {
	Color bool
}

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
)

func (r *TerminalReporter) Format() string { return "terminal" }

func (r *TerminalReporter) Report(result *types.Result, w io.Writer) error {
	if len(result.Findings) == 0 {
		fmt.Fprintf(w, "\n%s✓ No issues found%s\n", r.c(colorGreen), r.c(colorReset))
		r.printTrustScore(result.TrustScore, w)
		return nil
	}

	// Group findings by file, preserving order
	byFile := make(map[string][]types.Finding)
	var fileOrder []string
	for _, f := range result.Findings {
		if _, exists := byFile[f.File]; !exists {
			fileOrder = append(fileOrder, f.File)
		}
		byFile[f.File] = append(byFile[f.File], f)
	}

	for _, file := range fileOrder {
		findings := byFile[file]
		fmt.Fprintf(w, "\n%s%s%s\n", r.c(colorBold), file, r.c(colorReset))

		for _, f := range findings {
			sev := r.fmtSeverity(f.Severity)
			fmt.Fprintf(w, "  %d:%d  %s  %s  %s%s%s\n",
				f.Line, f.Column,
				sev, f.Message,
				r.c(colorDim), f.RuleID, r.c(colorReset))

			if f.Suggestion != "" {
				fmt.Fprintf(w, "         %s💡 %s%s\n",
					r.c(colorCyan), f.Suggestion, r.c(colorReset))
			}
		}
	}

	fmt.Fprintln(w)
	r.printSummary(result, w)
	r.printTrustScore(result.TrustScore, w)

	return nil
}

func (r *TerminalReporter) fmtSeverity(s types.Severity) string {
	switch s {
	case types.SeverityError:
		return fmt.Sprintf("%s✗ error  %s", r.c(colorRed), r.c(colorReset))
	case types.SeverityWarning:
		return fmt.Sprintf("%s⚠ warning%s", r.c(colorYellow), r.c(colorReset))
	case types.SeverityInfo:
		return fmt.Sprintf("%sℹ info   %s", r.c(colorCyan), r.c(colorReset))
	default:
		return "  unknown"
	}
}

func (r *TerminalReporter) printSummary(result *types.Result, w io.Writer) {
	var errors, warnings, infos int
	for _, f := range result.Findings {
		switch f.Severity {
		case types.SeverityError:
			errors++
		case types.SeverityWarning:
			warnings++
		case types.SeverityInfo:
			infos++
		}
	}

	var parts []string
	if errors > 0 {
		parts = append(parts, fmt.Sprintf("%s%d error(s)%s", r.c(colorRed), errors, r.c(colorReset)))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%s%d warning(s)%s", r.c(colorYellow), warnings, r.c(colorReset)))
	}
	if infos > 0 {
		parts = append(parts, fmt.Sprintf("%s%d info(s)%s", r.c(colorCyan), infos, r.c(colorReset)))
	}

	fmt.Fprintf(w, "  %s · %d files scanned · %d rules applied\n",
		strings.Join(parts, " · "), result.FilesScanned, result.RulesApplied)
}

func (r *TerminalReporter) printTrustScore(ts types.TrustScore, w io.Writer) {
	var scoreColor string
	switch {
	case ts.Score >= 80:
		scoreColor = colorGreen
	case ts.Score >= 60:
		scoreColor = colorYellow
	default:
		scoreColor = colorRed
	}

	fmt.Fprintf(w, "\n  %sTrust Score: %s%d/%d (%s)%s\n\n",
		r.c(colorBold), r.c(scoreColor),
		ts.Score, ts.MaxScore, ts.Grade, r.c(colorReset))
}

func (r *TerminalReporter) c(code string) string {
	if !r.Color {
		return ""
	}
	return code
}
