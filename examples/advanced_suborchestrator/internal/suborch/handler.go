package suborch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/entur/go-orchestrator"
	"github.com/kaptinlin/jsonschema"
)

const defaultTimeout = 10 * time.Second
const defaultDialerTimeout = 5 * time.Second

// -----------------------
// Sub-Orchestrator Manifest Handler (Car)
// -----------------------

// Airplane Manifest ---V
type AirplaneManifest struct {
	orchestrator.ManifestHeader
	Metadata AirplaneManifestMetadata `json:"metadata" jsonschema:"required"`
	Spec     AirplaneManifestSpec     `json:"spec" jsonschema:"required"`
}

type AirplaneManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type AirplaneManifestSpec struct {
}

var AirplanManifestSchema = jsonschema.FromStruct[AirplaneManifest]()

// Airplane Manifest Handler ---V
type AirplaneManifestHandler struct{
	client *http.Client
}

func (h *AirplaneManifestHandler) APIVersion() orchestrator.APIVersion {
	return "orchestrator.entur.io/vehicle/v1" // Which Manifest version this handler operates on
}

func (h *AirplaneManifestHandler) Kind() orchestrator.Kind {
	return "Airplane" // Which Manifest Kind this handler operates on
}


func (so *AirplaneManifestHandler) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest AirplaneManifest
	var err error 

	result := AirplanManifestSchema.ValidateJSON(req.Manifest.New)
	if !result.IsValid() {
		for path, msg := range result.GetDetailedErrors() {
			err = errors.Join(err, fmt.Errorf("%s: %s", path, msg))
		}
	} else {
		err = AirplanManifestSchema.Unmarshal(&manifest, req.Manifest.New)
	}

	if err != nil {
		r.Fail(fmt.Sprintf("Manifest is invalid:\n%s", err.Error()))
	} else {
		// Store parsed manifest on ctx
		orchestrator.Ctx(ctx).Set("manifest", manifest)
	}

	return nil
} 

func (so *AirplaneManifestHandler) MiddlewareAfter(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)

	return nil
} 

func (h *AirplaneManifestHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	manifest, ok := orchestrator.Ctx(ctx).Get("manifest").(AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	manifest, ok := orchestrator.Ctx(ctx).Get("manifest").(AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	manifest, ok := orchestrator.Ctx(ctx).Get("manifest").(AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

func NewAirplaneManifestHandler() *AirplaneManifestHandler {
	return &AirplaneManifestHandler{
		client: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: defaultDialerTimeout,
				}).Dial,
			},
		},
	}
}


// -----------------------
// Sub-Orchestrator Manifest Handler (Airplane)
// -----------------------
