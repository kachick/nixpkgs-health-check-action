package runner

import (
	"context"
	"errors"
	"testing"

	"github.com/kachick/nixpkgs-health-check-action/internal/config"
	"github.com/kachick/nixpkgs-health-check-action/internal/update"
)

func TestRunRespectsSkipsAndSortsOutput(t *testing.T) {
	cfg := &config.Config{
		Packages: map[string]config.PackageConfig{
			"alpha": {
				Hydra: config.ServiceConfig{Skip: config.SkipValue{Skip: true, Reason: "known"}},
			},
		},
	}

	hydraCalled := map[string]bool{}
	hydraRun := func(_ context.Context, pname string) ([]byte, error) {
		hydraCalled[pname] = true
		return []byte(`{"1/` + pname + `": {"still_succeed": [{"name": "` + pname + `-1.0", "job_name": "` + pname + `.x86_64-linux", "arch": "x86_64-linux", "status": "Succeeded"}]}}`), nil
	}
	updateRun := func(_ context.Context, pname string, _ string) (string, int, error) {
		if pname == "beta" {
			return "[ERROR] boom", 1, nil
		}
		return "[OK]", 0, nil
	}

	rep, err := Run(context.Background(), []string{"beta", "alpha", "beta"}, Options{
		Config:       cfg,
		RunHydra:     true,
		RunUpdate:    true,
		Concurrency:  2,
		HydraRunner:  hydraRun,
		UpdateRunner: updateRun,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(rep.Packages) != 2 {
		t.Fatalf("got %d packages, want 2", len(rep.Packages))
	}
	if rep.Packages[0].Pname != "alpha" || rep.Packages[1].Pname != "beta" {
		t.Errorf("packages not sorted: %+v", rep.Packages)
	}
	if hydraCalled["alpha"] {
		t.Errorf("alpha hydra check should have been skipped")
	}
	if !rep.Packages[0].HydraSkipped || rep.Packages[0].HydraSkipReason != "known" {
		t.Errorf("alpha skip state mismatch: %+v", rep.Packages[0])
	}
	if rep.Packages[1].Update == nil || rep.Packages[1].Update.Status != update.StatusError {
		t.Errorf("beta update should be error: %+v", rep.Packages[1].Update)
	}
	if !rep.HasErrors() || rep.ErrorCount() != 1 {
		t.Errorf("ErrorCount = %d, want 1", rep.ErrorCount())
	}
}

func TestRunPropagatesToolError(t *testing.T) {
	_, err := Run(context.Background(), []string{"foo"}, Options{
		RunHydra: true,
		HydraRunner: func(_ context.Context, _ string) ([]byte, error) {
			return nil, errors.New("hydra-check not found")
		},
	})
	if err == nil {
		t.Fatal("expected tool execution error to propagate")
	}
}
