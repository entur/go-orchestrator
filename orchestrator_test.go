package orchestrator_test

import (
	"context"
	"fmt"
	"os"

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
		Code:    orchestrator.ResultCodeSuccess,
		Changes: orchestrator.Changes{},
	}
	result.Changes.AddCreate("Created a thing")
	result.Changes.AddUpdate("Updated a thing")
	result.Changes.AddDelete("Deleted a thing")
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
	// ApiVersion, Kind and Metadata.Id is reqired
	manifest := ExampleManifest{
		ApiVersion: "orcestrator.entur.io/example/v1",
		Kind:       "Example",
		Metadata: orchestrator.Metadata{
			Id: "mything",
		},
	}

	so := ExampleSO{}
	project := "not-a-project" // os.Getenv("PROJECT_ID")
	client, _ := pubsub.NewClient(context.Background(), project)
	// TODO: disable time part so tests can pass
	handler := orchestrator.NewEventHandler(&so, client, orchestrator.WithCustomLogWriter(zerolog.ConsoleWriter{
		NoColor:      true,
		Out:          os.Stdout,
		PartsExclude: []string{"time"},
	}))

	event, _ := orchestrator.NewMockEvent(manifest, "plan")
	err := handler(context.Background(), *event)

	if err != nil {
		fmt.Println("HANDLER ERR:", err)
	}
	// Output:
	// 10:46AM INF UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgpDcmVhdGVkIGEgdGhpbmcKVXBkYXRlZDoKVXBkYXRlZCBhIHRoaW5nCkRlbGV0ZWQ6CkRlbGV0ZWQgYSB0aGluZwo= action=plan file_name= github_user_id=0 request_id=
	// HANDLER ERR: no topic set, cannot respond
}
