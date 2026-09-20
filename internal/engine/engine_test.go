package engine

import (
	"context"
	"errors"
	"github.com/SidharthSasikumar/ailint/internal/rule"
	"github.com/SidharthSasikumar/ailint/pkg/types"
	"testing"
)

type failingRule struct{ rule.Rule }

func (failingRule) ID() string          { return "broken" }
func (failingRule) Languages() []string { return nil }
func (failingRule) Check(context.Context, *types.FileContext) ([]types.Finding, error) {
	return nil, errors.New("analysis failed")
}
func TestIncompleteScanFails(t *testing.T) {
	e := New([]rule.Rule{failingRule{}}, nil, 2)
	if result, err := e.Run(context.Background(), []*types.FileContext{{}}); err == nil || result != nil {
		t.Fatal("failed analysis reported success")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := e.Run(ctx, nil); !errors.Is(err, context.Canceled) || result != nil {
		t.Fatal("cancelled analysis reported success")
	}
}
