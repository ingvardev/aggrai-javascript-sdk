// Package shared contains shared utilities and helpers.
package shared

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger is the application logger.
var Logger zerolog.Logger

func init() {
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// NewLogger creates a new logger with the given component name.
func NewLogger(component string) zerolog.Logger {
	return Logger.With().Str("component", component).Logger()
}
