package entrypoint

import (
	"os"

	"advanced_suborchestrator/internal/suborch"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/entur/go-logging"
	"github.com/entur/go-orchestrator"
)

// -----------------------
// Initialize Cloud Function
// -----------------------

func init() {
	// Read Config!
	projectID := os.Getenv("PROJECT_ID")
	functionEntrypoint := os.Getenv("FUNCTION_ENTRYPOINT")

	// Setup Logging!
	logger := logging.New()

	// Setup Sub-Orchestrator!
	mh := suborch.NewAirplaneManifestHandler()
	so := suborch.NewVehiclesSubOrch(projectID, mh)

	// Start Cloud Function!
	h := orchestrator.NewCloudEventHandler(so, orchestrator.WithCustomLogger(logger))
	functions.CloudEvent(functionEntrypoint, h)
}
