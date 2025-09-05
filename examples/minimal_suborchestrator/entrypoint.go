package suborch

import (
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/entur/go-orchestrator"
	"github.com/entur/go-orchestrator/examples/minimal_suborchestrator/internal/suborch"
)

// -----------------------
// Initialize Cloud Function
// -----------------------

func init() {
	// Read Config!
	projectID := os.Getenv("PROJECT_ID")
	functionEntrypoint := os.Getenv("FUNCTION_ENTRYPOINT")

	// Setup Sub-Orchestrator!
	mh := suborch.NewMyMinimalManifestHandler()
	so := suborch.NewMyMinimalSubOrch(projectID, mh)

	// Start Cloud Function!
	h := orchestrator.NewCloudEventHandler(so)
	functions.CloudEvent(functionEntrypoint, h)
}
