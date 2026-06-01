package atomic_test

import (
	"io"
	"testing"

	"github.com/adaouat/bifrost/internal/strategy"
	"github.com/adaouat/bifrost/internal/strategy/atomic"
)

func TestDeployerImplementsInterface(t *testing.T) {
	var _ strategy.Deployer = (*atomic.Deployer)(nil)
}

func TestNew_ReturnsDeployer(t *testing.T) {
	d := atomic.New(io.Discard, false, nil)
	if d == nil {
		t.Fatal("expected non-nil deployer")
	}
}
