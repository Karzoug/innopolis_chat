package main

import (
	"chat/internal/app"
	"context"
	"flag"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	app.Run(ctx)
}
