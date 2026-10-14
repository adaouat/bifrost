package cmdutil

import (
	"fmt"

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
