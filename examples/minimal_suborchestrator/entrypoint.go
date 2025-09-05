package entrypoint

import (
	"os"

	"minimal_suborchestrator/internal/suborch"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/entur/go-orchestrator"
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
