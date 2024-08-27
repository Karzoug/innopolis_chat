package domain

import (
	"encoding/json"
	"fmt"
	"time"
)

type MsgType string
type ReqType string
type DeliveryType string
type ID string

const (
	MsgTypeAdd MsgType = "add"
	MsgTypeUpd MsgType = "upd"
	MsgTypeDel MsgType = "del"

	ReqTypeNewChat ReqType = "new_chat"
	ReqTypeNewMsg  ReqType = "new_msg"

	DeliveryTypeNewMsg  DeliveryType = "new_msg"
	DeliveryTypeNewChat DeliveryType = "new_chat"
	DeliveryTypeError   DeliveryType = "error"
)

// Chat and Message is storage data
type Chat struct {
	UIDs     []ID      `json:"uid"`
	ChID     ID        `json:"ch_id"`
	Messages []Message `json:"messages"`
}

type Message struct {
	MsgID  ID        `json:"msg_id"`
	Body   string    `json:"body"`
	TDate  time.Time `json:"t_date"`
	FromID ID        `json:"from_id"`
}

// Request
// Data for Type = ReqTypeNewChat is NewChatRequest
// Data for Type = ReqTypeNewMsg is MessageChatRequest
type Request struct {
	Type ReqType         `json:"type"`
	Data json.RawMessage `json:"data"`
}

type NewChatRequest struct {
	UserIDs []ID `json:"users"`
}

// MessageChatRequest can support messages types:
// MsgTypeAdd is new msg in chat
// MsgTypeUpd is changing existing msg in chat
// MsgTypeDel is deleting msg from chat
type MessageChatRequest struct {
	Msg  string  `json:"msg"`
	Type MsgType `json:"type"`
	ChID ID      `json:"ch_id"`
}

// Delivery is msg that go to a client
// Data for Type = DeliveryTypeNewChat is chatID string
// Data for Type = DeliveryTypeNewMsg is MessageChatDelivery
// Data for Type = DeliveryTypeError is ErrorResponse
type Delivery struct {
	Type DeliveryType `json:"type"`
	Data interface{}  `json:"data"`
}

func (d *Delivery) UnmarshalJSON(data []byte) error {
	type MyDelivery Delivery

	aliasValue := &struct {
		*MyDelivery
		Data json.RawMessage `json:"data"`
	}{
		MyDelivery: (*MyDelivery)(d),
	}
	if err := json.Unmarshal(data, aliasValue); err != nil {
		return err
	}

	switch d.Type {
	case DeliveryTypeNewChat:
		var s string
		d.Data = s
		return json.Unmarshal(aliasValue.Data, &d.Data)
	case DeliveryTypeNewMsg:
		d.Data = &MessageChatDelivery{}
		return json.Unmarshal(aliasValue.Data, d.Data)
	case DeliveryTypeError:
		d.Data = &ErrorResponse{}
		return json.Unmarshal(aliasValue.Data, d.Data)
	default:
		return fmt.Errorf("unknown delivery type: %s", d.Type)
	}
}

// MessageChatDelivery can support messages types:
// MsgTypeAdd is new msg in chat
// MsgTypeUpd is changing existing msg in chat
// MsgTypeDel is deleting msg from chat
type MessageChatDelivery struct {
	Message
	Type MsgType `json:"type"`
	ChID ID      `json:"ch_id"`
}

// ErrorResponse can be supplemented with additional field "code"
// that client can easily handle
type ErrorResponse struct {
	Error string `json:"error"`
}
