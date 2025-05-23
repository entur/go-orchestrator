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

func (h ExampleSO) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	fmt.Println("Before it begins")
	if req.Sender.Type == orchestrator.SenderTypeUser {
		fmt.Println("#####")
		client := orchestrator.NewIAMLookupClient(req.Resources.IAM.Url)

		access, err := client.GCPUserHasRoleInProjects(ctx, req.Sender.Email, "your_so_role", "ent-someproject-dev")
		if err != nil {
			return err
		}

		if access == false {
			r.Done("You don't have access to ent-someproject-dev", false)
			return nil
		}
	}
	return nil
}

func (h ExampleSO) MiddlewareAfter(ctx context.Context, _ orchestrator.Request, _ *orchestrator.Result) error {
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

func (so *ExampleKindV1Handler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	r.Create("A thing")
	r.Update("A thing")
	r.Delete("A thing")
	r.Done("Plan all the things", true)
	return nil
}

func (so *ExampleKindV1Handler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plandestroy not implemented")
}

func (so *ExampleKindV1Handler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	r.Create("A thing")
	r.Update("A thing")
	r.Delete("A thing")
	r.Done("Plan all the things", true)
	return nil
}

func (so *ExampleKindV1Handler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
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

	iamServer := orchestrator.NewMockIAMLookupServer(
		orchestrator.WithPort(8001),
		orchestrator.WithUserProjectRoles(
			orchestrator.MockUserEmail,
			"ent-someproject-dev",
			[]string{"your_so_role"},
		),
	)
	iamResource, _ := iamServer.Serve()
	defer iamServer.Close()

	// Optional modifier of your mockevent
	mockEventModifier := func(r *orchestrator.Request) {
		r.Metadata.RequestID = "ExampleId"
		r.Resources.IAM = iamResource
	}

	event, _ := orchestrator.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, mockEventModifier)

	handler := orchestrator.NewEventHandler(so, orchestrator.WithCustomLogger(logger))
	// functions.CloudEvent("OrchestratorEvent", handler)

	err := handler(context.Background(), *event)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// DBG Created a new EventHandler
	// INF Ready to receive and process request gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request={"action":"plan","apiVersion":"orchestrator.entur.io/request/v1","manifest":{"new":{"apiVersion":"orchestation.entur.io/example/v1","kind":"Example","spec":{"name":"Test Name"}},"old":null},"metadata":{"requestId":"ExampleId"},"origin":{"fileName":"","repository":{"htmlUrl":""}},"resources":{"iamLookup":{"url":"http://localhost:8001"}},"responseTopic":"topic","sender":{"githubEmail":"mockuser@entur.io","githubId":0,"type":"user"}} gorch_request_id=ExampleId
	// DBG Found ManifestHandler orchestation.entur.io/example/v1 Example gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// DBG Executing MiddlewareBefore handler gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// Before it begins
	// #####
	// DBG Executing ManifestHandler orchestation.entur.io/example/v1 Example plan gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// DBG Executing MiddlewareAfter handler gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// After it's done
	// INF Ready to send response gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId gorch_response={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIEEgdGhpbmcKVXBkYXRlZDoKISBBIHRoaW5nCkRlbGV0ZWQ6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Encountered an internal error whilst responding to request error="no topic set, unable to respond" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// ERR Encountered error error="no topic set, unable to respond"
}
