package runner

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"github.com/kachick/nixpkgs-health-check-action/internal/config"
	"github.com/kachick/nixpkgs-health-check-action/internal/hydra"
	"github.com/kachick/nixpkgs-health-check-action/internal/update"
)

type Options struct {
	Config       *config.Config
	RunHydra     bool
	RunUpdate    bool
	Concurrency  int
	HydraRunner  hydra.Runner
	UpdateRunner update.Runner
}

type PackageReport struct {
	Pname            string
	HydraEnabled     bool
	HydraSkipped     bool
	HydraSkipReason  string
	Hydra            *hydra.CheckResult
	UpdateEnabled    bool
	UpdateSkipped    bool
	UpdateSkipReason string
	Update           *update.CheckResult
}

func (p PackageReport) ErrorCount() int {
	n := 0
	if p.Hydra != nil {
		n += len(p.Hydra.Errors)
	}
	if p.Update != nil && p.Update.Status == update.StatusError {
		n++
	}
	return n
}

func (p PackageReport) WarningCount() int {
	n := 0
	if p.Hydra != nil {
		n += len(p.Hydra.Warnings)
		if len(p.Hydra.StaleIgnore) > 0 {
			n++
		}
	}
	if p.Update != nil {
		if p.Update.Status == update.StatusWarning {
			n++
		}
		if p.Update.StaleIgnore {
			n++
		}
	}
	return n
}

type Report struct {
	Packages []PackageReport
}

func (r *Report) ErrorCount() int {
	n := 0
	for _, p := range r.Packages {
		n += p.ErrorCount()
	}
	return n
}

func (r *Report) WarningCount() int {
	n := 0
	for _, p := range r.Packages {
		n += p.WarningCount()
	}
	return n
}

func (r *Report) HasErrors() bool {
	return r.ErrorCount() > 0
}

func Run(ctx context.Context, pnames []string, opts Options) (*Report, error) {
	unique := make([]string, 0, len(pnames))
	for _, p := range pnames {
		if p != "" {
			unique = append(unique, p)
		}
	}
	slices.Sort(unique)
	unique = slices.Compact(unique)

	concurrency := opts.Concurrency
	if concurrency <= 0 {
		concurrency = 4
	}
	hydraRun := opts.HydraRunner
	if hydraRun == nil {
		hydraRun = hydra.DefaultRunner
	}
	updateRun := opts.UpdateRunner
	if updateRun == nil {
		updateRun = update.NewCommandRunner("")
	}

	results := make([]PackageReport, len(unique))
	errs := make([]error, len(unique))

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	for i, pname := range unique {
		wg.Add(1)
		go func(idx int, pkg string) {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				errs[idx] = ctx.Err()
				return
			}

			pr, err := checkPackage(ctx, pkg, opts.Config, opts.RunHydra, opts.RunUpdate, hydraRun, updateRun)
			if err != nil {
				errs[idx] = err
				return
			}
			results[idx] = pr
		}(i, pname)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return &Report{Packages: results}, nil
}

func checkPackage(
	ctx context.Context,
	pname string,
	cfg *config.Config,
	runHydra bool,
	runUpdate bool,
	hydraRun hydra.Runner,
	updateRun update.Runner,
) (PackageReport, error) {
	pr := PackageReport{
		Pname:         pname,
		HydraEnabled:  runHydra,
		UpdateEnabled: runUpdate,
	}

	if runHydra {
		if skip, reason := cfg.SkipHydra(pname); skip {
			pr.HydraSkipped = true
			pr.HydraSkipReason = reason
		} else {
			res, err := hydra.Check(
				ctx,
				hydraRun,
				pname,
				cfg.AllHydraIgnoreURLs(pname),
				cfg.PackageHydraIgnoreURLs(pname),
			)
			if err != nil {
				return pr, fmt.Errorf("hydra check for %q: %w", pname, err)
			}
			pr.Hydra = res
		}
	}

	if runUpdate {
		if skip, reason := cfg.SkipUpdate(pname); skip {
			pr.UpdateSkipped = true
			pr.UpdateSkipReason = reason
		} else {
			var cfgBytes []byte
			if cfg != nil {
				var err error
				cfgBytes, err = cfg.EncodeUpdateConfig(pname)
				if err != nil {
					return pr, err
				}
			}
			res, err := update.Check(
				ctx,
				updateRun,
				pname,
				cfgBytes,
				cfg.HasPackageUpdateIgnore(pname),
			)
			if err != nil {
				return pr, fmt.Errorf("nixpkgs-update check for %q: %w", pname, err)
			}
			pr.Update = res
		}
	}

	return pr, nil
}
