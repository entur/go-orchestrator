package orchestrator_test

import (
	"context"
	"fmt"

	"github.com/entur/go-logging"
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

func (s *ExampleSO) ProjectID() string {
	/* your project id */
	return ""
}

func (s *ExampleSO) Plan(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
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
func (s *ExampleSO) PlanDestroy(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("plandestroy not implemented")
}
func (s *ExampleSO) Apply(ctx context.Context, req orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	if req.Sender.Type == orchestrator.SenderTypeUser {
		client := req.Resources.IAM.ToClient()

		access, err := client.UserHasRoleOnProjects(ctx, req.Sender.Email, "your_so_role", "ent-someproject-dev")
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
func (s *ExampleSO) Destroy(context.Context, orchestrator.Request[ExampleManifest]) (orchestrator.Result, error) {
	return orchestrator.Result{}, fmt.Errorf("destroy not implemented")
}

func Example() {
	writer := zerolog.NewConsoleWriter()
	writer.NoColor = true
	writer.PartsExclude = []string{"timestamp"}
	logger := logging.New(logging.WithWriter(writer))

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
	handler := orchestrator.NewEventHandler(&so, orchestrator.WithCustomLogger(logger))

	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, func(r *orchestrator.Request[ExampleManifest]) { r.Metadata.RequestID = "ExampleId" })
	err := handler(context.Background(), *event)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// INF Response ready to send gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId gorch_response={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIEEgdGhpbmcKVXBkYXRlZDoKISBBIHRoaW5nCkRlbGV0ZWQ6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Could not respond error="no topic set, cannot respond" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// ERR Encountered error error="no topic set, cannot respond"
}
