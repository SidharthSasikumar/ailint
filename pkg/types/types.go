package types

import (
	"fmt"
	"strings"
)

// Severity levels for findings.
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
)

func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

func (s Severity) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// Finding is a single lint result.
type Finding struct {
	RuleID     string   `json:"rule_id"`
	RuleName   string   `json:"rule_name"`
	Severity   Severity `json:"severity"`
	File       string   `json:"file"`
	Line       int      `json:"line"`
	Column     int      `json:"column"`
	EndLine    int      `json:"end_line,omitempty"`
	EndColumn  int      `json:"end_column,omitempty"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
}

func (f Finding) String() string {
	return fmt.Sprintf("%s:%d:%d: [%s] %s: %s", f.File, f.Line, f.Column, f.Severity, f.RuleID, f.Message)
}

// FileContext holds content and metadata for a file under analysis.
type FileContext struct {
	Path     string
	Content  []byte
	Language string
	Lines    []string
}

// NewFileContext splits content into lines and sets the language.
func NewFileContext(path string, content []byte, language string) *FileContext {
	return &FileContext{
		Path:     path,
		Content:  content,
		Language: language,
		Lines:    strings.Split(string(content), "\n"),
	}
}

// Import is a parsed import statement.
type Import struct {
	Path  string // Full import path (e.g., "fmt", "github.com/pkg/errors")
	Alias string // Import alias, if any
	Line  int    // Line number in the source file
	Name  string // Short name (last segment of path)
}

// APICall is a parsed function/method call (used by future rules).
type APICall struct {
	Package  string
	Function string
	Line     int
	Column   int
	Raw      string
}

// TrustScore summarizes the overall scan result.
type TrustScore struct {
	Score    int    `json:"score"`
	MaxScore int    `json:"max_score"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
	Infos    int    `json:"infos"`
	Grade    string `json:"grade"`
}

// CalculateTrustScore derives a 0-100 score from findings.
func CalculateTrustScore(findings []Finding) TrustScore {
	ts := TrustScore{Score: 100, MaxScore: 100}
	for _, f := range findings {
		switch f.Severity {
		case SeverityError:
			ts.Errors++
			ts.Score -= 15
		case SeverityWarning:
			ts.Warnings++
			ts.Score -= 8
		case SeverityInfo:
			ts.Infos++
			ts.Score -= 3
		}
	}
	if ts.Score < 0 {
		ts.Score = 0
	}
	switch {
	case ts.Score >= 90:
		ts.Grade = "A"
	case ts.Score >= 80:
		ts.Grade = "B"
	case ts.Score >= 60:
		ts.Grade = "C"
	case ts.Score >= 40:
		ts.Grade = "D"
	default:
		ts.Grade = "F"
	}
	return ts
}

// Result is the complete output of an analysis run.
type Result struct {
	Findings     []Finding  `json:"findings"`
	TrustScore   TrustScore `json:"trust_score"`
	FilesScanned int        `json:"files_scanned"`
	RulesApplied int        `json:"rules_applied"`
}
