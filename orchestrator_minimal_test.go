package orchestrator_test

import (
	"context"
	"fmt"

	orchestrator "github.com/entur/go-orchestrator"
)

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

type MinimalSpec struct {
	Name string `json:"name"`
}

// apiVersion: orchestrator.entur.io/example/v1
// kind: Example
// spec: { name: Some Name }
type MinimalKind struct {
	orchestrator.ManifestHeader
	Spec MinimalSpec `json:"spec"`
}

type MinimalHandler struct{}

func (h *MinimalHandler) ApiVersion() orchestrator.ApiVersion {
	return "orchestrator.entur.io/example/v1"
}
func (h *MinimalHandler) Kind() orchestrator.Kind { return "Example" }

func (so *MinimalHandler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
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
	r.Create("A thing")
	r.Update("A thing")
	r.Delete("A thing")
	r.Done("Applied all the things", true)
	return nil
}

func (so *MinimalHandler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

func NewMinimalSO(projectID string) *MinimalSO {
	return &MinimalSO{
		projectID: projectID,
		handlers: []orchestrator.ManifestHandler{
			&MinimalHandler{},
		},
	}
}

// func init() {
// handler := orchestrator.NewEventHandler(so)
// functions.CloudEvent("OrchestratorEvent", orchestrator.NewEventHandler(so))
// }
func ExampleMinimalSO() {

	so := NewMinimalSO("mysoproject")

	manifest := MinimalKind{
		Spec: MinimalSpec{
			Name: "Test Name",
		},
		ManifestHeader: orchestrator.ManifestHeader{
			ApiVersion: so.handlers[0].ApiVersion(),
			Kind:       so.handlers[0].Kind(),
		},
	}

	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan)

	handler := orchestrator.NewEventHandler(so)

	err := handler(context.Background(), *event)

	if err != nil {
		fmt.Println(err)
	}
	// Output:
	// no topic set, unable to respond
}
