package orchestrator_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
	"github.com/entur/go-orchestrator/resources"
)

type ExampleSpecV1 struct {
	Name string `json:"name"`
}

type ExampleMetadataV1 struct {
	ID string `json:"id"`
}

type ExampleManifestV1 struct {
	orchestrator.ManifestHeader
	Metadata ExampleMetadataV1 `json:"metadata"`
	Spec     ExampleSpecV1     `json:"spec"`
}

type ExampleManifestV1Handler struct {
	/* you can have some internal state here */
}

func (h *ExampleManifestV1Handler) ApiVersion() orchestrator.ApiVersion {
	return "orchestrator.entur.io/example/v1"
}

func (h *ExampleManifestV1Handler) Kind() orchestrator.Kind {
	return "Example"
}

func (h *ExampleManifestV1Handler) MiddlewareBefore(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	logger := logging.Ctx(ctx)

	logger.Info().Msg("After Orchestrator middleware executes, but before manifest handler executes")

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

func (h *ExampleManifestV1Handler) PlanDestroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
	return fmt.Errorf("plandestroy not implemented")
}

func (h *ExampleManifestV1Handler) Apply(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
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

func (h *ExampleManifestV1Handler) Destroy(ctx context.Context, req orchestrator.Request, r *orchestrator.Result) error {
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
	logger := logging.Ctx(ctx)

	logger.Info().Msg("Before it begins")
	if req.Origin.Repository.Visibility != orchestrator.RepositoryVisbilityPublic {
		r.Done("This sub-orchestrator only accepts manifests in public repositories", false)
		return nil
	}

	if req.Sender.Type == orchestrator.SenderTypeUser {
		logger.Info().Msg("#####")
		client, err := resources.NewIAMClient(ctx, req.Resources.IAM.Url)
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
	cache := orchestrator.Ctx(ctx)
	cache.Set("cache_key", "something something!")

	return nil
}

func (so ExampleSO) MiddlewareAfter(ctx context.Context, _ orchestrator.Request, res *orchestrator.Result) error {
	logger := logging.Ctx(ctx)
	logger.Info().Msg("Auditing this thing")

	cache := orchestrator.Ctx(ctx)
	value := cache.Get("cache_key")
	if str, ok := value.(string); ok {
		logger.Info().Msgf("Got value from cache: %s", str)
	}

	logger.Info().Msg("After it's done")
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
// Complex Sub-Orchestrator Example
// -----------------------

func Example() {
	// Usually you would setup the sub-orchestrator inside an init function like so:
	//
	// 	func init() {
	//			handler := orchestrator.NewEventHandler(so)
	//	    	functions.CloudEvent("OrchestratorEvent", handler)
	//	}
	//
	// However, here we are configuring and executing it as part of an example test.

	logger := logging.New(
		logging.WithWriter(
			logging.NewConsoleWriter(
				logging.WithNoColor(),
				logging.WithNoTimestamp(),
			),
		),
	)

	iamServer, _ := resources.NewMockIAMServer(
		resources.WithPort(8001),
		resources.WithUserProjectRoles(
			orchestrator.DefaultMockUserEmail,
			"ent-someproject-dev",
			[]string{"your_so_role"},
		),
	)

	iamServer.Start()
	defer iamServer.Stop()

	so := NewExampleSO("mysoproject")
	handler := orchestrator.NewCloudEventHandler(so, orchestrator.WithCustomLogger(logger))

	manifest := ExampleManifestV1{
		ManifestHeader: orchestrator.ManifestHeader{
			ApiVersion: so.handlers[0].ApiVersion(),
			Kind:       so.handlers[0].Kind(),
		},
		Spec: ExampleSpecV1{
			Name: "Test Name",
		},
		Metadata: ExampleMetadataV1{
			ID: "manifestid",
		},
	}
	e, _ := orchestrator.NewMockCloudEvent(manifest, orchestrator.WithIAMEndpoint(iamServer.Url()))

	err := handler(context.Background(), *e)
	if err != nil {
		logger.Error().Err(err).Msg("Encountered error")
	}

	// Output:
	// DBG Created a new CloudEventHandler
	// DBG Processing request gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request={"action":"plan","apiVersion":"orchestrator.entur.io/request/v1","manifest":{"new":{"apiVersion":"orchestrator.entur.io/example/v1","kind":"Example","metadata":{"id":"manifestid"},"spec":{"name":"Test Name"}},"old":null},"metadata":{"requestId":"mockid"},"origin":{"fileName":"","repository":{"defaultBranch":"main","fullName":"","htmlUrl":"","id":0,"name":"","visibility":"public"}},"resources":{"iamLookup":{"url":"http://localhost:8001"}},"responseTopic":"mocktopic","sender":{"githubEmail":"mockuser@entur.io","githubId":0,"type":"user"}} gorch_request_id=mockid
	// DBG Found ManifestHandler (orchestrator.entur.io/example/v1, Example) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Executing Orchestrator MiddlewareBefore (mysoproject) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF Before it begins gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF ##### gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Unable to discover idtoken credentials, defaulting to http.Client for IAM gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Executing ManifestHandler MiddlewareBefore (orchestrator.entur.io/example/v1, Example, plan) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF After Orchestrator middleware executes, but before manifest handler executes gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Executing ManifestHandler (orchestrator.entur.io/example/v1, Example, plan) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Executing Orchestrator MiddlewareAfter (mysoproject) gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF Auditing this thing gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF Got value from cache: something something! gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// INF After it's done gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid
	// DBG Sending response gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid gorch_response={"apiVersion":"orchestrator.entur.io/request/v1","metadata":{"requestId":"mockid"},"output":"UGxhbiBhbGwgdGhlIHRoaW5ncwpDcmVhdGU6CisgQSB0aGluZwpVcGRhdGU6CiEgQSB0aGluZwpEZWxldGU6Ci0gQSB0aGluZwo=","result":"success"}
	// ERR Encountered an internal error whilst responding to request error="rpc error: code = NotFound desc = Resource not found (resource=mocktopic)." gorch_action=plan gorch_file_name= gorch_github_user_id=0 gorch_request_id=mockid gorch_result_creations=["A thing"] gorch_result_deletions=["A thing"] gorch_result_summary="Plan all the things" gorch_result_updates=["A thing"]
	// ERR Encountered error error="rpc error: code = NotFound desc = Resource not found (resource=mocktopic)."
}
