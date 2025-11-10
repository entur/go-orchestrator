package suborch

import (
	"context"

	"github.com/entur/go-orchestrator"
)

// -----------------------
// Vehicles Sub-Orchestrator
// -----------------------

type VehicleSubOrch struct {
	handlers  []orchestrator.ManifestHandler
}

func (so *VehicleSubOrch) Handlers() []orchestrator.ManifestHandler {
	return so.handlers
}

func (so *VehicleSubOrch) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	return nil
} 

func (so *VehicleSubOrch) MiddlewareAfter(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	return nil
} 

func NewVehiclesSubOrch(handlers ...orchestrator.ManifestHandler) *VehicleSubOrch {
	return &VehicleSubOrch{
		handlers:  handlers,
	}
}