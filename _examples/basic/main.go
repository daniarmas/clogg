package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/daniarmas/clogg"
)

func main() {
	ctx := context.Background()
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := clogg.New(clogg.LoggerConfig{
		BufferSize: 2,
		Handler:    handler,
	})
	defer logger.Shutdown()

	for i := 0; i < 500; i++ {
		msg := fmt.Sprintf("processing item %d", i)
		logger.Info(ctx, "processing item", []slog.Attr{
			slog.String("error", msg),
		}...)
	}

}
