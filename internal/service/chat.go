package service

import (
	"chat/internal/domain"
	"chat/internal/repository/chatdb"
	"chat/internal/service/pools"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

var chats chatdb.DB

func Init(chatDB chatdb.DB) {
	chats = chatDB
}

func NewMessage(msgReq domain.MessageChatRequest, fromID domain.ID) error {
	msg := domain.Message{
		MsgID:  domain.ID(uuid.New().String()),
		Body:   msgReq.Msg,
		TDate:  time.Now(),
		FromID: fromID,
	}

	if err := chats.AddMessage(msgReq.ChID, msg); err != nil {
		return err
	}

	users, err := chats.GetChatUsers(msgReq.ChID)
	if err != nil {
		return err
	}

	delivery := domain.Delivery{
		Type: domain.DeliveryTypeNewMsg,
		Data: domain.MessageChatDelivery{
			Message: msg,
			Type:    msgReq.Type,
			ChID:    msgReq.ChID,
		},
	}

	for _, userID := range users {
		if userID != fromID {
			pools.Users.Send(userID, delivery)
		}
	}
	return nil
}

func NewChat(uids []domain.ID) domain.ID {
	id := chats.AddChat(uids)

	log.Info().
		Str("id", string(id)).
		Msg("new chat created")

	return id
}
