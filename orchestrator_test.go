package orchestrator_test

import (
	"context"
	"fmt"

	"github.com/entur/go-logging"
	orchestrator "github.com/entur/go-orchestrator"
	"github.com/rs/zerolog"
)

type ExampleSOManifest struct {
	ApiVersion orchestrator.ApiVersion `json:"apiVersion"`
	Kind       orchestrator.Kind       `json:"kind"`
	Metadata   orchestrator.Metadata   `json:"metadata"`
}

type ExampleSO struct {
	/* you can have some internal state here */
	projectID string
}

func (so *ExampleSO) ProjectID() string {
	/* your project id */
	return so.projectID
}

func (so *ExampleSO) Plan(ctx context.Context, req orchestrator.Request[ExampleSOManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{
		Summary: "Plan all the things",
		Success: true,
		Creations: []string{
			"A thing",
		},
		Updates: []string{
			"A thing",
		},
		Deletions: []string{
			"A thing",
		},
	}, nil
}

func (so *ExampleSO) PlanDestroy(ctx context.Context, req orchestrator.Request[ExampleSOManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("plandestroy not implemented")
}

func (so *ExampleSO) Apply(ctx context.Context, req orchestrator.Request[ExampleSOManifest]) (orchestrator.Result, error) {
	if req.Sender.Type == orchestrator.SenderTypeUser {
		client := req.Resources.IAM.ToClient()

		access, err := client.GCPUserHasRoleInProjects(ctx, req.Sender.Email, "your_so_role", "ent-someproject-dev")
		if err != nil {
			return orchestrator.Result{}, err
		}

		if access == false {
			// Forbidden, so tell the user why!
			return orchestrator.Result{
				Summary: "You don't have access to ent-someproject-dev",
				Success: false,
			}, nil
		}
	}

	return orchestrator.Result{
		Summary: "Apply all the things",
		Success: true,
		Creations: []string{
			"A thing",
		},
		Updates: []string{
			"A thing",
		},
		Deletions: []string{
			"A thing",
		},
	}, nil
}

func (so *ExampleSO) Destroy(ctx context.Context, req orchestrator.Request[ExampleSOManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("destroy not implemented")
}

func NewExampleSO(projectID string) *ExampleSO {
	return &ExampleSO{
		projectID: projectID,
	}
}

// -----------------------
// Minimal Sub-Orchestrator Example
// -----------------------

func Example() {
	writer := zerolog.NewConsoleWriter()
	writer.NoColor = true
	writer.PartsExclude = []string{"timestamp"}
	logger := logging.New(logging.WithWriter(writer))

	// Just an example manifest, here is where you specify _your_ sub-orchestrator
	// ApiVersion, Kind and Metadata.ID is required
	manifest := ExampleSOManifest{
		ApiVersion: "orcestrator.entur.io/example/v1",
		Kind:       "Example",
		Metadata: orchestrator.Metadata{
			ID: "mything",
		},
	}

	// Optional modifier of your mockevent
	mockEventModifier := func(r *orchestrator.Request[ExampleSOManifest]) {
		r.Metadata.RequestID = "ExampleId"
	}

	so := NewExampleSO("mysoproject")
	handler := orchestrator.NewEventHandler(so, orchestrator.WithCustomLogger(logger))
	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, mockEventModifier)
	err := handler(context.Background(), *event)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// INF Response ready to send gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId gorch_response={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIEEgdGhpbmcKVXBkYXRlZDoKISBBIHRoaW5nCkRlbGV0ZWQ6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Could not respond error="no topic set, cannot respond" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// ERR Encountered error error="no topic set, cannot respond"
}
