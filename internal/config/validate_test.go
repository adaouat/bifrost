package config_test

import (
	"testing"

	"github.com/adaouat/bifrost/internal/config"
)

func TestValidate_OK(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
	}
	errs := config.Validate(m)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_MissingReleasesRoot(t *testing.T) {
	m := &config.MergedConfig{SharedRoot: "/var/www/shared"}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingSharedRoot(t *testing.T) {
	m := &config.MergedConfig{ReleasesRoot: "/var/www/releases"}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
}

func TestValidate_MissingBoth(t *testing.T) {
	m := &config.MergedConfig{}
	errs := config.Validate(m)
	if len(errs) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidate_UnsupportedStrategy(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
		Strategy:     "docker",
	}
	errs := config.Validate(m)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for unsupported strategy, got %d: %v", len(errs), errs)
	}
}

func TestValidate_AtomicStrategyOK(t *testing.T) {
	m := &config.MergedConfig{
		ReleasesRoot: "/var/www/releases",
		SharedRoot:   "/var/www/shared",
		Strategy:     "atomic",
	}
	if errs := config.Validate(m); len(errs) != 0 {
		t.Errorf("expected no errors for atomic, got: %v", errs)
	}
}
