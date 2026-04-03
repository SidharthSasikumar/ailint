package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/SidharthSasikumar/ailint/internal/config"
	"github.com/SidharthSasikumar/ailint/internal/engine"
	"github.com/SidharthSasikumar/ailint/internal/reporter"
	"github.com/SidharthSasikumar/ailint/internal/rule"
	"github.com/SidharthSasikumar/ailint/internal/scanner"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

const banner = `
    _    ___ _     _       _   
   / \  |_ _| |   (_)_ __ | |_ 
  / _ \  | || |   | | '_ \| __|
 / ___ \ | || |___| | | | | |_ 
/_/   \_\___|_____|_|_| |_|\__|
`

func main() {
	var (
		configPath  string
		format      string
		noColor     bool
		showVersion bool
		workers     int
		showBanner  bool
	)

	flag.StringVar(&configPath, "config", ".ailint.yaml", "Path to configuration file")
	flag.StringVar(&configPath, "c", ".ailint.yaml", "Path to configuration file (shorthand)")
	flag.StringVar(&format, "format", "", "Output format: terminal, json, sarif")
	flag.StringVar(&format, "f", "", "Output format (shorthand)")
	flag.BoolVar(&noColor, "no-color", false, "Disable colored output")
	flag.BoolVar(&showVersion, "version", false, "Show version information")
	flag.BoolVar(&showVersion, "v", false, "Show version (shorthand)")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "Number of parallel workers")
	flag.IntVar(&workers, "j", runtime.NumCPU(), "Number of parallel workers (shorthand)")
	flag.BoolVar(&showBanner, "banner", true, "Show banner")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintf(os.Stderr, "  Static analysis for AI-generated code\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  ailint [flags] [path]\n\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  ailint                    Scan current directory\n")
		fmt.Fprintf(os.Stderr, "  ailint ./src              Scan specific directory\n")
		fmt.Fprintf(os.Stderr, "  ailint -f json .          Output as JSON\n")
		fmt.Fprintf(os.Stderr, "  ailint -f sarif . > r.sarif  Output as SARIF\n")
		fmt.Fprintf(os.Stderr, "  ailint -c custom.yaml .   Use custom config\n")
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("ailint %s (commit: %s, built: %s)\n", version, commit, buildTime)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to load config: %v\n", err)
		os.Exit(1)
	}

	// CLI flags override config
	if format != "" {
		cfg.Output.Format = format
	}
	if noColor {
		cfg.Output.Color = false
	}

	// Determine scan root
	root := "."
	if args := flag.Args(); len(args) > 0 {
		root = args[0]
	}

	// Validate root exists
	if info, err := os.Stat(root); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %q is not a valid directory\n", root)
		os.Exit(1)
	}

	// Print banner for terminal output
	if showBanner && cfg.Output.Format == "terminal" {
		fmt.Fprint(os.Stderr, banner)
		fmt.Fprintf(os.Stderr, "  v%s\n\n", version)
	}

	// Initialize rules
	var rules []rule.Rule
	if cfg.Rules.PhantomImports.Enabled {
		rules = append(rules, rule.NewPhantomImports(root))
	}
	if cfg.Rules.SecretLeaks.Enabled {
		rules = append(rules, rule.NewSecretLeaks(cfg.Rules.SecretLeaks.EntropyThreshold))
	}
	if cfg.Rules.StalePatterns.Enabled {
		rules = append(rules, rule.NewStalePatterns())
	}

	if len(rules) == 0 {
		fmt.Fprintln(os.Stderr, "warning: no rules enabled")
		os.Exit(0)
	}

	// Initialize reporter
	var rep reporter.Reporter
	switch cfg.Output.Format {
	case "json":
		rep = &reporter.JSONReporter{}
	case "sarif":
		rep = &reporter.SARIFReporter{Version: version}
	default:
		rep = &reporter.TerminalReporter{Color: cfg.Output.Color}
	}

	// Scan files
	s := scanner.New(root, cfg.Scan.Exclude)
	files, err := s.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to scan files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "No supported files found")
		os.Exit(0)
	}

	// Run analysis
	eng := engine.New(rules, rep, workers)
	result, err := eng.Run(context.Background(), files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: analysis failed: %v\n", err)
		os.Exit(1)
	}

	// Write report
	if err := rep.Report(result, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: failed to write report: %v\n", err)
		os.Exit(1)
	}

	// Exit with appropriate code
	if result.TrustScore.Errors > 0 {
		os.Exit(1)
	}
	if cfg.Trust.Enabled && result.TrustScore.Score < cfg.Trust.Thresholds.Pass {
		os.Exit(1)
	}
}
