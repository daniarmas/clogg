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
	logger := clogg.GetLogger(clogg.LoggerConfig{
		BufferSize: 100,
		Handler:    handler,
	})
	defer logger.Shutdown()

	for i := 0; i < 500; i++ {
		msg := fmt.Sprintf("processing item %d", i)
		clogg.Info(ctx, "processing item", []clogg.Attr{
			clogg.String("error", msg),
		}...)
	}

}
