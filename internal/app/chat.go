package app

import (
	"chat/internal/handler"
	"chat/internal/repository/cache"
	"chat/internal/server"
	"chat/internal/service"
	"chat/pkg/authclient"
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

func Run(ctx context.Context) {
	wg := sync.WaitGroup{}

	chatDB, err := cache.ChatCacheInit(ctx, &wg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize chat database")
	}
	// initialize service
	service.Init(chatDB)
	authclient.Init("localhost:8000")

	go func() {
		log.Info().Str("address", "localhost:8001").Msg("chat server started")
		err := server.Run("localhost", "8001", http.HandlerFunc(handler.HandleHTTPReq))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server run")
		}
	}()

	<-ctx.Done()

	if err := server.Shutdown(); err != nil {
		log.Fatal().Err(err).Msg("server shutdown")
	}
	wg.Wait()
	log.Info().Msg("chat server stopped")
}
