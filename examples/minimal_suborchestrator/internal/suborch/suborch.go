package suborch

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/entur/go-orchestrator"
)

// -----------------------
// Sub-Orchestrator
// -----------------------

type MyMinimalSubOrch struct {
	projectID string
	handlers  []orchestrator.ManifestHandler
}

func (so *MyMinimalSubOrch) ProjectID() string {
	return so.projectID
}

func (so *MyMinimalSubOrch) Handlers() []orchestrator.ManifestHandler {
	return so.handlers
}

func NewMyMinimalSubOrch(projectID string, handlers ...orchestrator.ManifestHandler) *MyMinimalSubOrch {
	return &MyMinimalSubOrch{
		projectID: projectID,
		handlers:  handlers,
	}
}

// -----------------------
// Sub-Orchestrator Manifest Handlers
// -----------------------

// Your Manifest Definition --V
type MyMinimalManifest struct {
	orchestrator.ManifestHeader
	Metadata MyMinimalManifestMetadata `json:"metadata"`
	Spec     MyMinimalManifestSpec     `json:"spec"`
}

type MyMinimalManifestMetadata = orchestrator.ManifestMetadata // Default metadata definition, but you can use your own

type MyMinimalManifestSpec struct {
	Your   string   `json:"your"`
	Values []string `json:"values"`
	Here   int      `json:"here"`
}

// Your Manifest Handler ---V
type MyMinimalManifestHandler struct{}

func (h *MyMinimalManifestHandler) APIVersion() orchestrator.APIVersion {
	return "orchestrator.entur.io/MyMinimalSubOrch/v1" // Which Manifest version this handler operates on
}

func (h *MyMinimalManifestHandler) Kind() orchestrator.Kind {
	return "MyMinimalManifest" // Which Manifest Kind this handler operates on
}

func (h *MyMinimalManifestHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MyMinimalManifest

	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return err
	}

	if manifest.Spec.Here >= 10 {
		r.Fail(".spec.here value must be less than 10!")
		return nil
	}

	r.Create("A thing")
	r.Update("A thing")
	r.Succeed("Plan all the things")

	return nil
}

func (h *MyMinimalManifestHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MyMinimalManifest

	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return err
	}

	r.Delete("Some message")

	return nil
}

func (h *MyMinimalManifestHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MyMinimalManifest

	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return err
	}

	if manifest.Spec.Here >= 10 {
		r.Fail(".spec.here value must be less than 10!")
		return nil
	}

	r.Create("A thing")
	r.Update("A thing")
	r.Succeed("Applied all the things")

	return nil
}

func (h *MyMinimalManifestHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

func NewMyMinimalManifestHandler() *MyMinimalManifestHandler {
	return &MyMinimalManifestHandler{}
}
