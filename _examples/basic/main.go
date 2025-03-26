package main

import (
	"context"
	"log/slog"

	"github.com/daniarmas/clogg"
)

func main() {
	ctx := context.Background()
	logger := clogg.New(2)

	for i := 0; i < 2; i++ {
		logger.Info(ctx, "processing item", []slog.Attr{
			slog.String("error", "this is a test"),
		}...)
	}

	logger.Shutdown()
}
