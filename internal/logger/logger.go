package logger

import (
	"io"
	"log/slog"
	"os"
)

func New(level slog.Level, output io.Writer) *slog.Logger {
	if output == nil {
		output = os.Stdout
	}

	return slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	}))
}
