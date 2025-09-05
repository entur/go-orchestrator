package entrypoint

import (
	"database/sql"
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

	// Setup DB!
	db, err := sql.Open("", "")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to setup db")
	}

	// Setup Sub-Orchestrator!
	//cmh := suborch.NewCarManifestHandler(db)
	amh := suborch.NewAirplaneManifestHandler(db)
	so := suborch.NewVehiclesSubOrch(projectID, amh)

	// Start Cloud Function!
	h := orchestrator.NewCloudEventHandler(so, orchestrator.WithCustomLogger(logger))
	functions.CloudEvent(functionEntrypoint, h)
}
