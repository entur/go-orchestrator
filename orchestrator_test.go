package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
	"github.com/entur/go-orchestrator/event"
	"github.com/entur/go-orchestrator/resources"
)

type ExampleSpecV1 struct {
	Name string `json:"name"`
}

type ExampleManifestV1 struct {
	orchestrator.ManifestHeader
	Spec ExampleSpecV1 `json:"spec"`
}

type ExampleManifestV1Handler struct {
	/* you can have some internal state here */
}

func (h *ExampleManifestV1Handler) ApiVersion() orchestrator.ApiVersion {
	return "orchestation.entur.io/example/v1"
}

func (h *ExampleManifestV1Handler) Kind() orchestrator.Kind { 
	return "Example" 
}

func (h *ExampleManifestV1Handler) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	fmt.Println("After Orhcestrator middleware executes, but before manifest handler executes")
	return nil
}

func (h *ExampleManifestV1Handler) Plan(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest ExampleManifestV1
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

func (so *ExampleManifestV1Handler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plandestroy not implemented")
}

func (so *ExampleManifestV1Handler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	var manifest ExampleManifestV1
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

func (so *ExampleManifestV1Handler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("destroy not implemented")
}

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

func (so *ExampleSO) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	fmt.Println("Before it begins")
	if req.Sender.Type == orchestrator.SenderTypeUser {
		fmt.Println("#####")
		client, err := resources.NewIAMLookupClient(ctx, req.Resources.IAM.Url)
		if err != nil {
			return err
		}

		access, err := client.GCPUserHasRoleInProjects(ctx, req.Sender.Email, "your_so_role", "ent-someproject-dev")
		if err != nil {
			return err
		}

		if access == false {
			r.Done("You don't have access to ent-someproject-dev", false)
			return nil
		}
	}

	// The cache is shared between middlewares and handlers!
	cache := orchestrator.CtxCache(ctx)
	cache.Set("cache_key", "something something!")

	return nil
}

func (h ExampleSO) MiddlewareAfter(ctx context.Context, _ orchestrator.Request, res *orchestrator.Result) error {
	logger := logging.Ctx(ctx)
	logger.Info().Msg("Auditing this thing")

	cache := orchestrator.CtxCache(ctx)
	value := cache.Get("cache_key")
	if str, ok := value.(string); ok {
		fmt.Printf("Got value from cache: %s\n", str)
	}

	fmt.Println("After it's done")
	return nil
}

func NewExampleSO(projectID string) *ExampleSO {
	return &ExampleSO{
		projectID: projectID,
		handlers: []orchestrator.ManifestHandler{
			&ExampleManifestV1Handler{},
		},
	}
}

// -----------------------
// Minimal Sub-Orchestrator Example
// -----------------------

func Example() {
	writer := logging.NewConsoleWriter(logging.WithNoColor(), logging.WithNoTimestamp())
	logger := logging.New(logging.WithWriter(writer))

	so := NewExampleSO("mysoproject")

	manifest := ExampleManifestV1{
		Spec: ExampleSpecV1{
			Name: "Test Name",
		},
		ManifestHeader: orchestrator.ManifestHeader{
			ApiVersion: so.handlers[0].ApiVersion(),
			Kind:       so.handlers[0].Kind(),
		},
	}

	iamServer, _ := resources.NewMockIAMLookupServer(
		resources.WithPort(8001),
		resources.WithUserProjectRoles(
			event.MockUserEmail,
			"ent-someproject-dev",
			[]string{"your_so_role"},
		),
	)

	iamServer.Start()
	defer iamServer.Stop()

	// Optional modifier of your mockevent
	mockEventModifier := func(r *orchestrator.Request) {
		r.Metadata.RequestID = "ExampleId"
		r.Resources.IAM.Url = iamServer.Url()
	}

	e, _ := event.NewMockEvent(manifest, orchestrator.SenderTypeUser, orchestrator.ActionPlan, mockEventModifier)
	handler := event.NewEventHandler(so, event.WithCustomLogger(logger))
	// functions.CloudEvent("OrchestratorEvent", handler)

	err := handler(context.Background(), *e)

	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}
	// Output:
	// DBG Created a new EventHandler
	// INF Received and processing request gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request={"action":"plan","apiVersion":"orchestrator.entur.io/request/v1","manifest":{"new":{"apiVersion":"orchestation.entur.io/example/v1","kind":"Example","spec":{"name":"Test Name"}},"old":null},"metadata":{"requestId":"ExampleId"},"origin":{"fileName":"","repository":{"htmlUrl":""}},"resources":{"iamLookup":{"url":"http://localhost:8001"}},"responseTopic":"topic","sender":{"githubEmail":"mockuser@entur.io","githubId":0,"type":"user"}} gorch_request_id=ExampleId
	// DBG Found ManifestHandler (orchestation.entur.io/example/v1, Example) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// DBG Executing Orchestrator (mysoproject) MiddlewareBefore gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// Before it begins
	// #####
	// DBG Unable to discover idtoken credentials, defaulting to http.Client for IAMLookup gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// DBG Executing ManifestHandler (orchestation.entur.io/example/v1, Example, plan) MiddlewareBefore gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// After Orhcestrator middleware executes, but before manifest handler executes
	// DBG Executing ManifestHandler (orchestation.entur.io/example/v1 Example plan) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// DBG Executing Orchestrator (mysoproject) MiddlewareAfter gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// INF Auditing this thing gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// Got value from cache: something something!
	// After it's done
	// INF Sending response gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId gorch_response={"apiVersion":"orchestrator.entur.io/response/v1","metadata":{"requestId":"ExampleId"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGVkOgorIEEgdGhpbmcKVXBkYXRlZDoKISBBIHRoaW5nCkRlbGV0ZWQ6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Encountered an internal error whilst responding to request error="no topic set, unable to respond" gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=ExampleId
	// ERR Encountered error error="no topic set, unable to respond"
}
