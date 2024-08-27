package main

import (
	"context"
	"os/signal"
	"syscall"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(),
		syscall.SIGTERM,
		syscall.SIGKILL,
		syscall.SIGINT)

	logger, closeLogger := newLogger()
	defer closeLogger()

	cfg := mustLoadConfig()

	app := newApp(cfg, logger)
	app.run(ctx)
}
