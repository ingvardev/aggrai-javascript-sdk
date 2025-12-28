package shared
// Package shared contains shared utilities and helpers.
package shared

import (
	"os"
	"time"



















}	return Logger.With().Str("component", component).Logger()func NewLogger(component string) zerolog.Logger {// NewLogger creates a new logger with the given component name.}	Logger = zerolog.New(output).With().Timestamp().Caller().Logger()	}		TimeFormat: time.RFC3339,		Out:        os.Stdout,	output := zerolog.ConsoleWriter{func init() {var Logger zerolog.Logger// Logger is the application logger.)	"github.com/rs/zerolog"
