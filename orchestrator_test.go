package orchestrator_test

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	orchestrator "github.com/entur/go-orchestrator"
	"github.com/rs/zerolog"
)

type ExampleManifest struct {
	ApiVersion orchestrator.ApiVersion `json:"apiVersion"`
	Kind       orchestrator.Kind       `json:"kind"`
	Metadata   orchestrator.Metadata   `json:"metadata"`
}

type ExampleSO struct {
	/* you can have some internal state here */
}

func (s *ExampleSO) Plan(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	result := orchestrator.Result{
		Summary: "Plan all the things",
		Success:    true,
	}
	result.Creations = append(result.Creations, "Created a thing")
	result.Updates = append(result.Updates, "Updated a thing")
	result.Deletions = append(result.Deletions, "Created a thing")
	return result, nil
}
func (s *ExampleSO) PlanDestroy(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("plandestroy not implemented")
}
func (s *ExampleSO) Apply(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("apply not implemented")
}
func (s *ExampleSO) Destroy(context.Context, orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("destroy not implemented")
}

func ExampleOrchestrator() {
	// Just an example manifest, here is where you specify _your_ sub-orchestrator
	// ApiVersion, Kind and Metadata.ID is required
	manifest := ExampleManifest{
		ApiVersion: "orcestrator.entur.io/example/v1",
		Kind:       "Example",
		Metadata: orchestrator.Metadata{
			ID: "mything",
		},
	}

	so := ExampleSO{}
	project := "not-a-project" // os.Getenv("PROJECT_ID")
	client, _ := pubsub.NewClient(context.Background(), project)
	writer := zerolog.NewConsoleWriter()
	writer.NoColor = true
	writer.PartsExclude = []string{"timestamp"}
	handler := orchestrator.NewEventHandler(&so, client, orchestrator.WithCustomLogWriter(writer))

	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan)
	err := handler(context.Background(), *event)

	if err != nil {
		fmt.Println("HANDLER ERR:", err)
	}
	// Output:
	// INF Response ready to send gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id= gorch_response={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"request_id":""},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIENyZWF0ZWQgYSB0aGluZwpVcGRhdGVkOgohIFVwZGF0ZWQgYSB0aGluZwpEZWxldGVkOgotIENyZWF0ZWQgYSB0aGluZwo=","result":"success"}
	// ERR Could not respond error="no topic set, cannot respond" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=
	// HANDLER ERR: no topic set, cannot respond
}
