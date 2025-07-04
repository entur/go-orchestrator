package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/entur/go-orchestrator"
	"github.com/entur/go-orchestrator/event"
)

type MinimalMetadata struct {
	ID string `json:"id"`
}

type MinimalSpec struct {
	Name string `json:"name"`
}

// apiVersion: orchestrator.entur.io/example/v1
// kind: Example
// metadata: { id: Some Id }
// spec: { name: Some Name }
type MinimalManifest struct {
	orchestrator.ManifestHeader
	Metadata ExampeMetadataV1 `json:"metadata"`
	Spec MinimalSpec `json:"spec"`
}

type MinimalHandler struct {
	/* you can have some internal state here */
}

func (h *MinimalHandler) ApiVersion() orchestrator.ApiVersion {
	return "orchestrator.entur.io/example/v1"
}

func (h *MinimalHandler) Kind() orchestrator.Kind {
	return "Example"
}

func (so *MinimalHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MinimalManifest
	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return err
	}

	r.Create("A thing")
	r.Update("A thing")
	r.Delete("A thing")
	r.Done("Plan all the things", true)
	return nil
}

func (so *MinimalHandler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plandestroy not implemented")
}

func (so *MinimalHandler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest MinimalManifest
	err := json.Unmarshal(req.Manifest.New, &manifest)
	if err != nil {
		return err
	}

	r.Create("A thing")
	r.Update("A thing")
	r.Delete("A thing")
	r.Done("Applied all the things", true)
	return nil
}

func (so *MinimalHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

type MinimalSO struct {
	projectID string
	handlers  []orchestrator.ManifestHandler
}

func (so *MinimalSO) ProjectID() string {
	return so.projectID
}

func (so *MinimalSO) Handlers() []orchestrator.ManifestHandler {
	return so.handlers
}

func NewMinimalSO(projectID string) *MinimalSO {
	return &MinimalSO{
		projectID: projectID,
		handlers: []orchestrator.ManifestHandler{
			&MinimalHandler{},
		},
	}
}

//	func init() {
//	    handler := orchestrator.NewEventHandler(so)
//	    functions.CloudEvent("OrchestratorEvent", handler)
//	}
func ExampleMinimalSO() {

	so := NewMinimalSO("mysoproject")

	manifest := MinimalManifest{
		Spec: MinimalSpec{
			Name: "Test Name",
		},
		ManifestHeader: orchestrator.ManifestHeader{
			ApiVersion: so.handlers[0].ApiVersion(),
			Kind:       so.handlers[0].Kind(),
		},
	}

	e, _ := event.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan)
	handler := event.NewEventHandler(so)

	err := handler(context.Background(), *e)

	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// no topic set, unable to respond
}
