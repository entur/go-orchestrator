package suborch

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/entur/go-orchestrator"
	"github.com/kaptinlin/jsonschema"
)

// -----------------------
// Sub-Orchestrator Manifest Handler (Airplane)
// -----------------------

// Airplane Manifest
type AirplaneManifest struct {
	orchestrator.ManifestHeader
	Metadata AirplaneManifestMetadata `json:"metadata" jsonschema:"required"`
	Spec     AirplaneManifestSpec     `json:"spec" jsonschema:"required"`
}

type AirplaneManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type AirplaneManifestSpec struct {
	Model string `json:"model" jsonschema:"required"`
	Wingspan float64 `json:"wingspanMeters" jsonschema:"required,minimum=1,maximum=500"`
	Passengers int `json:"numberOfPassengers" jsonschema:"required,minimum=0,maximum=500"`
}

var AirplaneManifestSchema = jsonschema.FromStruct[AirplaneManifest]()

// Airplane Manifest Handler
type AirplaneManifestHandler struct{
	db *sql.DB
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

	// Handle validation of manifest
	result := AirplaneManifestSchema.ValidateJSON(req.Manifest.New)
	if !result.IsValid() {
		for path, msg := range result.GetDetailedErrors() {
			err = errors.Join(err, fmt.Errorf("%s: %s", path, msg))
		}
	} else {
		err = AirplaneManifestSchema.Unmarshal(&manifest, req.Manifest.New)
	}

	// If manifest is invalid, report it as a failure to the user.
	// Else, save the parsed manifest for processing in later handlers
	if err != nil {
		r.Fail(fmt.Sprintf("Manifest is invalid:\n%s", err.Error()))
	} else {
		orchestrator.Ctx(ctx).Set("manifest", &manifest) // Store parsed manifest on ctx
	}

	return nil
} 

func (so *AirplaneManifestHandler) MiddlewareAfter(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)

	return nil
} 

func (h *AirplaneManifestHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *AirplaneManifestHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*AirplaneManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func NewAirplaneManifestHandler(db *sql.DB) *AirplaneManifestHandler {
	return &AirplaneManifestHandler{
		db: db,
	}
}

// -----------------------
// Sub-Orchestrator Manifest Handler (Car)
// -----------------------

// Car Manifest
type CarManifest struct {
	orchestrator.ManifestHeader
	Metadata CarManifestMetadata `json:"metadata" jsonschema:"required"`
	Spec     CarManifestSpec     `json:"spec" jsonschema:"required"`
}

type CarManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type CarManifestSpec struct {
	Model string `json:"model" jsonschema:"required"`
	Wheels float64 `json:"numberOfWheels" jsonschema:"required,minimum=1,maximum=99"`
	Passengers int `json:"numberOfPassengers" jsonschema:"required,minimum=0,maximum=100"`
}

var CarManifestSchema = jsonschema.FromStruct[AirplaneManifest]()

// Car Manifest Handler
type CarManifestHandler struct{
	db *sql.DB
}

func (h *CarManifestHandler) APIVersion() orchestrator.APIVersion {
	return "orchestrator.entur.io/vehicle/v1" // Which Manifest version this handler operates on
}

func (h *CarManifestHandler) Kind() orchestrator.Kind {
	return "Car" // Which Manifest Kind this handler operates on
}

func (so *CarManifestHandler) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest AirplaneManifest
	var err error 

	// Handle validation of manifest
	result := CarManifestSchema.ValidateJSON(req.Manifest.New)
	if !result.IsValid() {
		for path, msg := range result.GetDetailedErrors() {
			err = errors.Join(err, fmt.Errorf("%s: %s", path, msg))
		}
	} else {
		err = CarManifestSchema.Unmarshal(&manifest, req.Manifest.New)
	}

	// If manifest is invalid, report it as a failure to the user.
	// Else, save the parsed manifest for processing in later handlers
	if err != nil {
		r.Fail(fmt.Sprintf("Manifest is invalid:\n%s", err.Error()))
	} else {
		orchestrator.Ctx(ctx).Set("manifest", &manifest) // Store parsed manifest on ctx
	}

	return nil
} 

func (so *CarManifestHandler) MiddlewareAfter(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)

	return nil
} 

func (h *CarManifestHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*CarManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *CarManifestHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*CarManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *CarManifestHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*CarManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func (h *CarManifestHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	//logger := logging.Ctx(ctx)
	_, ok := orchestrator.Ctx(ctx).Get("manifest").(*CarManifest)
	if !ok {
		return fmt.Errorf("couldn't retrieve parsed manifest")
	}

	return nil
}

func NewCarManifestHandler(db *sql.DB) *AirplaneManifestHandler {
	return &AirplaneManifestHandler{
		db: db,
	}
}