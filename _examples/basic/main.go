package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/daniarmas/clogg"
)

func main() {
	ctx := context.Background()
	logger := clogg.New(2)

	for i := 0; i < 1000; i++ {
		msg := fmt.Sprintf("processing item %d", i)
		logger.Info(ctx, "processing item", []slog.Attr{
			slog.String("error", msg),
		}...)
	}

	logger.Shutdown()
}
