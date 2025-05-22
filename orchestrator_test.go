package orchestrator_test

import (
	"context"
	"fmt"

	"github.com/entur/go-logging"
	orchestrator "github.com/entur/go-orchestrator"
	"github.com/rs/zerolog"
)

type ExampleSO struct {
	/* you can have some internal state here */
	projectID string
	handlers  []orchestrator.ManifestHandler
}

func (so *ExampleSO) ProjectID() string {
	return so.projectID
}

func (so *ExampleSO) Handlers() []orchestrator.ManifestHandler {
	return so.handlers
}

type ExampleSpecV1 struct {
	Name string `json:"name"`
}
type ExampleKindV1 struct {
	orchestrator.ManifestHeader
	Spec ExampleSpecV1 `json:"spec"`
}
type ExampleKindV1Handler struct{}

func (h *ExampleKindV1Handler) ApiVersion() orchestrator.ApiVersion {
	return "orchestation.entur.io/example/v1"
}
func (h *ExampleKindV1Handler) Kind() orchestrator.Kind { return "Example" }

func (so *ExampleKindV1Handler) Plan(ctx context.Context, req orchestrator.Request, res *orchestrator.ResponseResult) error {
	res.Create("A thing")
	res.Update("A thing")
	res.Delete("A thing")
	res.Done("Plan all the things", true)
	return nil
}

func (so *ExampleKindV1Handler) PlanDestroy(ctx context.Context, req orchestrator.Request, res *orchestrator.ResponseResult) error {
	return fmt.Errorf("plandestroy not implemented")
}

func (so *ExampleKindV1Handler) Apply(ctx context.Context, req orchestrator.Request, res *orchestrator.ResponseResult) error {
	if req.Sender.Type == orchestrator.SenderTypeUser {
		client := orchestrator.NewIAMLookupClient(req.Resources.IAM.Url)

		access, err := client.GCPUserHasRoleInProjects(ctx, req.Sender.Email, "your_so_role", "ent-someproject-dev")
		if err != nil {
			return err
		}

		if access == false {
			res.Done("You don't have access to ent-someproject-dev", false)
			return nil
		}
	}
	res.Create("A thing")
	res.Update("A thing")
	res.Delete("A thing")
	res.Done("Plan all the things", true)
	return nil
}

func (so *ExampleKindV1Handler) Destroy(ctx context.Context, req orchestrator.Request, res *orchestrator.ResponseResult) error {
	return fmt.Errorf("destroy not implemented")
}

func NewExampleSO(projectID string) *ExampleSO {
	return &ExampleSO{
		projectID: projectID,
		handlers: []orchestrator.ManifestHandler{
			&ExampleKindV1Handler{},
		},
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

	// Optional modifier of your mockevent
	mockEventModifier := func(r *orchestrator.Request) {
		r.Metadata.RequestID = "ExampleId"
	}

	so := NewExampleSO("mysoproject")
	manifest := ExampleKindV1{
		Spec: ExampleSpecV1{
			Name: "Test Name",
		},
		ManifestHeader: orchestrator.ManifestHeader{
			ApiVersion: so.handlers[0].ApiVersion(),
			Kind:       so.handlers[0].Kind(),
		},
	}

	handler := orchestrator.NewEventHandler(so, orchestrator.WithCustomLogger(logger))
	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, mockEventModifier)
	err := handler(context.Background(), *event)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// INF Created a new EventHandler
	// INF Handling request gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId req={"action":"plan","apiVersion":"orchestrator.entur.io/request/v1","manifest":{"new":{"apiVersion":"orchestation.entur.io/example/v1","kind":"Example","spec":{"name":"Test Name"}},"old":null},"metadata":{"requestId":"ExampleId"},"origin":{"fileName":"","repository":{"htmlUrl":""}},"resources":{"iamLookup":{"url":""}},"responseTopic":"topic","sender":{"githubEmail":"","githubId":0,"type":"user"}}
	// INF Found handler for orchestation.entur.io/example/v1 Example gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// INF Performed plan on orchestation.entur.io/example/v1 Example gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// INF Got response gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId res={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIEEgdGhpbmcKVXBkYXRlZDoKISBBIHRoaW5nCkRlbGV0ZWQ6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Encountered error error="no topic set, cannot respond"
}
