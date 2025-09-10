package suborch

import (
	"context"

	"github.com/entur/go-orchestrator"
)

// -----------------------
// Vehicles Sub-Orchestrator
// -----------------------

type VehicleSubOrch struct {
	projectID string
	handlers  []orchestrator.ManifestHandler
}

func (so *VehicleSubOrch) ProjectID() string {
	return so.projectID
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

func NewVehiclesSubOrch(projectID string, handlers ...orchestrator.ManifestHandler) *VehicleSubOrch {
	return &VehicleSubOrch{
		projectID: projectID,
		handlers:  handlers,
	}
}