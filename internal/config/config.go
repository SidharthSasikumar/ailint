package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the complete AILint configuration.
type Config struct {
	Version int          `yaml:"version"`
	Rules   RulesConfig  `yaml:"rules"`
	Output  OutputConfig `yaml:"output"`
	Scan    ScanConfig   `yaml:"scan"`
	Trust   TrustConfig  `yaml:"trust_score"`
}

// RulesConfig holds configuration for all rules.
type RulesConfig struct {
	PhantomImports  RuleConfig        `yaml:"phantom-imports"`
	SecretLeaks     SecretLeaksConfig `yaml:"secret-leaks"`
	StalePatterns   RuleConfig        `yaml:"stale-patterns"`
	ZombieAPIs      RuleConfig        `yaml:"zombie-apis"`
	ConventionDrift RuleConfig        `yaml:"convention-drift"`
	LicenseRisk     RuleConfig        `yaml:"license-risk"`
	ComplexityBombs RuleConfig        `yaml:"complexity-bombs"`
}

// RuleConfig holds common configuration for a single rule.
type RuleConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Severity  string   `yaml:"severity"`
	Languages []string `yaml:"languages,omitempty"`
}

// SecretLeaksConfig extends RuleConfig with secret-specific options.
type SecretLeaksConfig struct {
	RuleConfig       `yaml:",inline"`
	EntropyThreshold float64  `yaml:"entropy_threshold"`
	Patterns         []string `yaml:"patterns,omitempty"`
}

// OutputConfig controls how results are displayed.
type OutputConfig struct {
	Format  string `yaml:"format"`
	Color   bool   `yaml:"color"`
	Verbose bool   `yaml:"verbose"`
}

// ScanConfig controls which files are analyzed.
type ScanConfig struct {
	Paths    []string `yaml:"paths"`
	Exclude  []string `yaml:"exclude"`
	DiffOnly bool     `yaml:"diff_only"`
}

// TrustConfig controls trust score behavior.
type TrustConfig struct {
	Enabled    bool            `yaml:"enabled"`
	Thresholds TrustThresholds `yaml:"thresholds"`
}

// TrustThresholds defines pass/warn score thresholds.
type TrustThresholds struct {
	Pass int `yaml:"pass"`
	Warn int `yaml:"warn"`
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Rules: RulesConfig{
			PhantomImports: RuleConfig{Enabled: true, Severity: "error"},
			SecretLeaks: SecretLeaksConfig{
				RuleConfig:       RuleConfig{Enabled: true, Severity: "error"},
				EntropyThreshold: 4.5,
			},
			StalePatterns:   RuleConfig{Enabled: true, Severity: "warning"},
			ZombieAPIs:      RuleConfig{Enabled: false, Severity: "warning"},
			ConventionDrift: RuleConfig{Enabled: false, Severity: "info"},
			LicenseRisk:     RuleConfig{Enabled: false, Severity: "warning"},
			ComplexityBombs: RuleConfig{Enabled: false, Severity: "info"},
		},
		Output: OutputConfig{
			Format: "terminal",
			Color:  true,
		},
		Scan: ScanConfig{
			Paths:   []string{"."},
			Exclude: []string{"vendor/", "node_modules/", ".git/", "dist/", "build/"},
		},
		Trust: TrustConfig{
			Enabled:    true,
			Thresholds: TrustThresholds{Pass: 80, Warn: 60},
		},
	}
}

// Load reads configuration from the given YAML file, merging with defaults.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	return cfg, nil
}
