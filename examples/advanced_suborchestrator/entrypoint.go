package entrypoint

import (
	"database/sql"
	"os"

	"advanced_suborchestrator/internal/suborch"

	_ "modernc.org/sqlite"

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
	db, err := sql.Open("sqlite", "mock")
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to setup db")
	}

	// Setup Sub-Orchestrator!
	amh := suborch.NewAirplaneManifestHandler(db)
	cmh := suborch.NewCarManifestHandler(db)
	so := suborch.NewVehiclesSubOrch(projectID, amh, cmh)

	// Start Cloud Function!
	h := orchestrator.NewCloudEventHandler(so, orchestrator.WithCustomLogger(logger))
	functions.CloudEvent(functionEntrypoint, h)
}
