package reporter

import (
	"encoding/json"
	"io"

	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// SARIFReporter writes SARIF 2.1.0 output for GitHub Code Scanning.
type SARIFReporter struct {
	Version string
}

func (r *SARIFReporter) Format() string { return "sarif" }

func (r *SARIFReporter) Report(result *types.Result, w io.Writer) error {
	ruleMap := map[string]types.Finding{}
	for _, f := range result.Findings {
		if _, ok := ruleMap[f.RuleID]; !ok {
			ruleMap[f.RuleID] = f
		}
	}

	var rules []sarifRule
	for id, f := range ruleMap {
		rules = append(rules, sarifRule{
			ID:               id,
			Name:             f.RuleName,
			ShortDescription: sarifText{Text: f.RuleName},
			DefaultConfig:    sarifRuleConfig{Level: sarifLevel(f.Severity)},
		})
	}

	var results []sarifResult
	for _, f := range result.Findings {
		msg := f.Message
		if f.Suggestion != "" {
			msg += " | Suggestion: " + f.Suggestion
		}
		col := f.Column
		if col < 1 {
			col = 1
		}
		results = append(results, sarifResult{
			RuleID:  f.RuleID,
			Level:   sarifLevel(f.Severity),
			Message: sarifText{Text: msg},
			Locations: []sarifLocation{
				{
					PhysicalLocation: sarifPhysicalLocation{
						ArtifactLocation: sarifArtifactLocation{URI: f.File},
						Region:           sarifRegion{StartLine: f.Line, StartColumn: col},
					},
				},
			},
		})
	}

	version := r.Version
	if version == "" {
		version = "0.1.0"
	}

	log := sarifLog{
		Version: "2.1.0",
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "ailint",
						Version:        version,
						InformationURI: "https://github.com/SidharthSasikumar/ailint",
						Rules:          rules,
					},
				},
				Results: results,
			},
		},
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

func sarifLevel(s types.Severity) string {
	switch s {
	case types.SeverityError:
		return "error"
	case types.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}

// SARIF 2.1.0 schema types

type sarifLog struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ShortDescription sarifText       `json:"shortDescription"`
	DefaultConfig    sarifRuleConfig `json:"defaultConfiguration"`
}

type sarifRuleConfig struct {
	Level string `json:"level"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
}
