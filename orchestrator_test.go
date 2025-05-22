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

func (h ExampleSO) MiddlewareBefore(ctx context.Context, req orchestrator.Request, res *orchestrator.ResponseResult) error {
	fmt.Println("Before it begins")
	if req.Sender.Type == orchestrator.SenderTypeUser {
		fmt.Println("#####")
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
	return nil
}

func (h ExampleSO) MiddlewareAfter(ctx context.Context, _ orchestrator.Request, _ *orchestrator.ResponseResult) error {
	fmt.Println("After it's done")
	return nil
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

	// Optional modifier of your mockevent
	mockEventModifier := func(r *orchestrator.Request) {
		r.Metadata.RequestID = "ExampleId"
	}
	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, mockEventModifier)

	handler := orchestrator.NewEventHandler(so, orchestrator.WithCustomLogger(logger))
	// functions.CloudEvent("OrchestratorEvent", handler)

	err := handler(context.Background(), *event)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// INF Created a new EventHandler
	// INF Handling request gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId req={"action":"plan","apiVersion":"orchestrator.entur.io/request/v1","manifest":{"new":{"apiVersion":"orchestation.entur.io/example/v1","kind":"Example","spec":{"name":"Test Name"}},"old":null},"metadata":{"requestId":"ExampleId"},"origin":{"fileName":"","repository":{"htmlUrl":""}},"resources":{"iamLookup":{"url":"example.com"}},"responseTopic":"topic","sender":{"githubEmail":"","githubId":0,"type":"user"}}
	// INF Found handler for orchestation.entur.io/example/v1 Example gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// INF Executing MiddlewareBefore gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// Before it begins
	// #####
	// ERR error="no client passed to request" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId gorch_result={}
	// INF Got response gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId res={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"QW4gaW50ZXJuYWwgZXJyb3Igb2NjdXJlZA==","result":"error"}
	// ERR Encountered error error="no topic set, cannot respond"
}
