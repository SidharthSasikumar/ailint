package engine

import (
	"context"
	"sync"

	"github.com/SidharthSasikumar/ailint/internal/reporter"
	"github.com/SidharthSasikumar/ailint/internal/rule"
	"github.com/SidharthSasikumar/ailint/pkg/types"
)

// Engine orchestrates file analysis across rules using a worker pool.
type Engine struct {
	rules    []rule.Rule
	reporter reporter.Reporter
	workers  int
}

// New creates an Engine with the given rules, reporter, and worker count.
func New(rules []rule.Rule, rep reporter.Reporter, workers int) *Engine {
	if workers <= 0 {
		workers = 4
	}
	return &Engine{
		rules:    rules,
		reporter: rep,
		workers:  workers,
	}
}

// Run analyzes all files and returns the aggregated result.
func (e *Engine) Run(ctx context.Context, files []*types.FileContext) (*types.Result, error) {
	var (
		mu       sync.Mutex
		findings []types.Finding
		wg       sync.WaitGroup
	)

	// Buffer all files into the channel
	fileCh := make(chan *types.FileContext, len(files))
	for _, f := range files {
		fileCh <- f
	}
	close(fileCh)

	// Fan out workers
	for i := 0; i < e.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for file := range fileCh {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				for _, r := range e.rules {
					// Check if rule applies to this file's language
					langs := r.Languages()
					if len(langs) > 0 {
						applicable := false
						for _, l := range langs {
							if l == file.Language {
								applicable = true
								break
							}
						}
						if !applicable {
							continue
						}
					}

					results, err := r.Check(ctx, file)
					if err != nil {
						continue
					}

					if len(results) > 0 {
						mu.Lock()
						findings = append(findings, results...)
						mu.Unlock()
					}
				}
			}
		}()
	}

	wg.Wait()

	return &types.Result{
		Findings:     findings,
		TrustScore:   types.CalculateTrustScore(findings),
		FilesScanned: len(files),
		RulesApplied: len(e.rules),
	}, nil
}
