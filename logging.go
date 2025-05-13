package orchestrator

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Configure zerolog for GCP logging
	zerolog.LevelFieldName = "severity"
	zerolog.LevelWarnValue = "WARNING"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Set log level
	level := os.Getenv("LOG_LEVEL")
	switch strings.ToLower(level) {
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	}
}

func NewLogger(w io.Writer) zerolog.Logger {
	if w != nil {
		return log.Output(w).With().Logger()
	}
	return log.With().Logger()
}
