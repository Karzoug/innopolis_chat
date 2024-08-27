package handler

import (
	"chat/internal/domain"
	"chat/pkg/authclient"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const HeaderAuthorization = "Authorization"
const HeaderUserID = "User-ID"
const HeaderUserRole = "User-Role"

var upgrader = websocket.Upgrader{
	HandshakeTimeout: time.Minute,
	ReadBufferSize:   1024,
	WriteBufferSize:  1024,
	WriteBufferPool:  &sync.Pool{},
	CheckOrigin: func(r *http.Request) bool {
		return true // Пропускаем любой запрос
	},
}

func HandleHTTPReq(resp http.ResponseWriter, req *http.Request) {
	defer func() {
		resp.Header().Set("Access-Control-Allow-Origin", "*")
		resp.Header().Add("Access-Control-Allow-Methods", "GET")
		resp.Header().Add("Access-Control-Allow-Methods", "OPTIONS")
	}()

	token := req.Header.Get(HeaderAuthorization)

	if token == "" {
		resp.WriteHeader(http.StatusUnauthorized)
		log.Debug().
			Str("method", req.Method).
			Str("token", token).
			Str("error", http.StatusText(http.StatusUnauthorized))
		return
	}

	userID, valid := authclient.ValidateToken(token)
	if !valid {
		resp.WriteHeader(http.StatusUnauthorized)
		log.Debug().
			Str("method", req.Method).
			Str("token", token).
			Str("error", http.StatusText(http.StatusUnauthorized))
		return
	}

	log.Info().Str("user ID", userID).Msg("user connected")

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		log.Error().Err(err).Msg("failed to upgrade conn")
		return
	}
	HandleWsConn(conn, domain.ID(userID))
}
