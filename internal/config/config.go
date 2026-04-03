package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the complete ailint configuration.
type Config struct {
	Version int          `yaml:"version"`
	Rules   RulesConfig  `yaml:"rules"`
	Output  OutputConfig `yaml:"output"`
	Scan    ScanConfig   `yaml:"scan"`
	Trust   TrustConfig  `yaml:"trust_score"`
}

// RulesConfig groups per-rule configuration.
type RulesConfig struct {
	PhantomImports  RuleConfig        `yaml:"phantom-imports"`
	SecretLeaks     SecretLeaksConfig `yaml:"secret-leaks"`
	StalePatterns   RuleConfig        `yaml:"stale-patterns"`
	ZombieAPIs      RuleConfig        `yaml:"zombie-apis"`
	ConventionDrift RuleConfig        `yaml:"convention-drift"`
	LicenseRisk     RuleConfig        `yaml:"license-risk"`
	ComplexityBombs RuleConfig        `yaml:"complexity-bombs"`
}

// RuleConfig is the common configuration for a rule.
type RuleConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Severity  string   `yaml:"severity"`
	Languages []string `yaml:"languages,omitempty"`
}

// SecretLeaksConfig adds entropy threshold to RuleConfig.
type SecretLeaksConfig struct {
	RuleConfig       `yaml:",inline"`
	EntropyThreshold float64  `yaml:"entropy_threshold"`
	Patterns         []string `yaml:"patterns,omitempty"`
}

// OutputConfig controls output formatting.
type OutputConfig struct {
	Format  string `yaml:"format"`
	Color   bool   `yaml:"color"`
	Verbose bool   `yaml:"verbose"`
}

// ScanConfig controls file discovery.
type ScanConfig struct {
	Paths    []string `yaml:"paths"`
	Exclude  []string `yaml:"exclude"`
	DiffOnly bool     `yaml:"diff_only"`
}

// TrustConfig controls trust score thresholds.
type TrustConfig struct {
	Enabled    bool            `yaml:"enabled"`
	Thresholds TrustThresholds `yaml:"thresholds"`
}

// TrustThresholds defines the pass/warn cutoffs.
type TrustThresholds struct {
	Pass int `yaml:"pass"`
	Warn int `yaml:"warn"`
}

// DefaultConfig returns production defaults.
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

// Load reads a YAML config file, falling back to defaults.
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
