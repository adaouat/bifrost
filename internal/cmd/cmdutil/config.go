package cmdutil

import (
	"fmt"
	"os"

	"github.com/adaouat/bifrost/internal/cmderr"
	"github.com/adaouat/bifrost/internal/config"
)

// ResolveMergedConfig resolves cfg to a MergedConfig for the given env/app. A
// flat config (no environments) skips the env/app requirement; otherwise both
// are required. The result is validated. Missing flags surface as plain errors
// (Usage exit code); validation failures carry the Config exit code.
func ResolveMergedConfig(cfg *config.Config, env, app string) (*config.MergedConfig, error) {
	var merged *config.MergedConfig
	if config.IsFlat(cfg) {
		merged = config.MergeFlat(cfg)
	} else {
		if env == "" {
			return nil, fmt.Errorf("--env (or --environment) is required")
		}
		if app == "" {
			return nil, fmt.Errorf("--app (or --application) is required")
		}
		m, err := config.Merge(cfg, env, app)
		if err != nil {
			return nil, err
		}
		merged = m
	}
	if errs := config.Validate(merged); len(errs) > 0 {
		return nil, &cmderr.ExitError{Code: cmderr.Config, Message: errs.Error()}
	}
	return merged, nil
}

// WriteTempFlatConfig generates a flat config for env/app and writes it to a
// temp file, returning the path and a cleanup func the caller must invoke.
func WriteTempFlatConfig(cfg *config.Config, env, app string) (path string, cleanup func(), err error) {
	f, err := os.CreateTemp("", "bifrost-config-*.yml")
	if err != nil {
		return "", nil, fmt.Errorf("creating temp config: %w", err)
	}
	cleanFn := func() { _ = os.Remove(f.Name()) }

	if err := config.GenerateFlatConfig(cfg, env, app, f); err != nil {
		cleanFn()
		_ = f.Close()
		return "", nil, fmt.Errorf("writing flat config: %w", err)
	}
	if err := f.Close(); err != nil {
		cleanFn()
		return "", nil, fmt.Errorf("closing temp config: %w", err)
	}
	return f.Name(), cleanFn, nil
}
