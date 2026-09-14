package main

import (
	"log/slog"
	"os"

	"github.com/ESilva15/TelemetryMockserver/cmd"
)

func init() {
	_ = setupLogger()
}

func setupLogger() error {
	output, err := os.OpenFile("./output.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}

	logger := slog.New(
		slog.NewTextHandler(output, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
	)

	slog.SetDefault(logger)

	return nil
}

func main() {
	cmd.Execute()
}
