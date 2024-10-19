package internal

import (
	"log/slog"
	"os"
)

// NewLogger creates a new logger.
// the loglevel is INFO if verbose is false, DEBUG otherwise.
func NewLogger(verbose bool) *slog.Logger {
	var level slog.Level
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}
