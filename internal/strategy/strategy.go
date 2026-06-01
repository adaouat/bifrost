package strategy

import (
	"context"

	"github.com/adaouat/bifrost/internal/config"
)

// DeployOptions carries the inputs for a single deployment.
type DeployOptions struct {
	Config      *config.MergedConfig
	Artifact    string
	ReleaseName string
	Env         string
	App         string
}

// Deployer executes a deployment for a specific strategy.
type Deployer interface {
	Deploy(ctx context.Context, opts DeployOptions) error
}
