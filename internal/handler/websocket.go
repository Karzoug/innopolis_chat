package handler

import (
	"chat/internal/domain"
	"chat/internal/service"
	"chat/internal/service/pools"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	writeWait  = 1 * time.Second
	pongWait   = 10 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

func HandleWsConn(conn *websocket.Conn, UID domain.ID) {
	defer func() {
		// closing the user channel and ending write goroutine
		if pools.Users.Delete(UID) {
			log.Info().Str("UID", string(UID)).Msg("conn closed")
			conn.Close()
		}
	}()

	ch := pools.Users.New(UID)

	// write to conn from channel
	go func() {
		ticker := time.NewTicker(pingPeriod)
		defer func() {
			ticker.Stop()
			if pools.Users.Delete(UID) {
				log.Info().Str("UID", string(UID)).Msg("conn closed")
				conn.Close()
			}
		}()
		for {
			select {
			case msg, ok := <-ch:
				if !ok {
					log.Debug().Str("UID", string(UID)).Msg("channel closed")
					return
				}
				log.Debug().Str("UID", string(UID)).Any("msg", msg).Msg("send message")

				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteJSON(msg); err != nil {
					handleWsError(err, UID)
					return
				}
			case <-ticker.C:
				log.Trace().Str("UID", string(UID)).Msg("send ping")
				conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					handleWsError(err, UID)
					return
				}
			}
		}
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		log.Trace().Str("from UID", string(UID)).Msg("got pong")
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	// read from conn
	for {
		typ, message, err := conn.ReadMessage()
		if err != nil {
			handleWsError(err, UID)
			return
		}
		log.Debug().
			Str("from UID", string(UID)).
			Int("type", typ).
			Str("message", string(message)).
			Msg("got message")
		switch typ {
		case websocket.TextMessage, websocket.BinaryMessage:
			var req domain.Request
			if err = json.Unmarshal(message, &req); err != nil {
				sendErrorResp(UID, err)
				continue
			}
			switch req.Type {
			case domain.ReqTypeNewChat:
				var newChatReq domain.NewChatRequest
				if err = json.Unmarshal(req.Data, &newChatReq); err != nil {
					sendErrorResp(UID, err)
					continue
				}
				chatid := service.NewChat(append(newChatReq.UserIDs, UID))
				sendResp(UID, domain.DeliveryTypeNewChat, chatid)

			case domain.ReqTypeNewMsg:
				var msg domain.MessageChatRequest
				if err = json.Unmarshal(req.Data, &msg); err != nil {
					sendErrorResp(UID, err)
					continue
				}

				switch msg.Type {
				case domain.MsgTypeAdd:
					if err := service.NewMessage(msg, UID); err != nil {
						sendErrorResp(UID, err)
						continue
					}
				}
			}
		case websocket.CloseMessage:
			return
		}
	}
}

func handleWsError(err error, uid domain.ID) {
	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway):
		log.Debug().Str("UID", string(uid)).Msg("websocket session closed by client")
	default:
		log.Debug().Str("UID", string(uid)).Err(err).Msg("error websocket message")
	}
}

func sendErrorResp(UID domain.ID, err error) {
	sendResp(UID, domain.DeliveryTypeError, domain.ErrorResponse{Error: err.Error()})
}

func sendResp(UID domain.ID, typ domain.DeliveryType, data interface{}) {
	var resp domain.Delivery
	resp.Type = typ
	resp.Data = data
	log.Debug().
		Str("UID", string(UID)).
		Any("resp", resp).
		Msg("send to channel")
	pools.Users.Send(UID, resp)
}
