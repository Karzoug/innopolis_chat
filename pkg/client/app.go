package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"

	"chat/internal/domain"
)

type state uint

const (
	menu state = iota
	createChat
	inChat
)

const returnText = "return"

type connectFunc func() (conn *websocket.Conn, closeFn func())

type app struct {
	curentState state
	chatID      string
	connectFn   connectFunc
	logger      zerolog.Logger
}

func newApp(cfg Config, logger zerolog.Logger) *app {
	return &app{
		connectFn: func() (conn *websocket.Conn, closeFn func()) {
			headers := make(http.Header)
			headers.Add("Authorization", cfg.Token)

			conn, resp, err := websocket.DefaultDialer.Dial("ws://"+cfg.Address, headers)
			if err != nil {
				if resp != nil && resp.StatusCode == http.StatusUnauthorized {
					fmt.Println("unauthorized: wrong token")
					os.Exit(1)
				}
				color.Red("failed to connect")
				logger.Fatal().Err(err).Msg("failed to connect")
			}
			return conn, func() {
				if err := conn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")); err != nil {
					logger.Error().Err(err).Msg("conn write close")
				}
			}
		},
		logger: logger,
	}
}

func (app *app) run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		switch app.curentState {
		case menu:
			if !app.menu() {
				return
			}
		case inChat:
			app.chat(ctx)
			app.curentState = menu
		}
		fmt.Print("\n")
	}
}

func (app *app) menu() bool {
	fmt.Println(`--- Основное меню ---
1. Создать новый чат с другим пользователем
2. Войти в чат с пользователем

Введите ваш выбор (для выхода введите exit или нажмите ctrl+C)`)

	var choice string
	fmt.Scanln(&choice)

	switch choice {
	case "1":
		var id string
		fmt.Println(`Введите ID пользователя, с которым вы бы хотели начать чат или введите return для выхода в предыдущее меню.`)
		fmt.Scanln(&id)
		switch id {
		case returnText:
		default:
			chID, err := app.requestNewChat(id)
			if err != nil {
				app.logger.Error().Err(err).Msg("failed to request new chat")
				color.Red("Не удалось создать ID чата")
				return true
			}
			fmt.Println("Ваш ID чата: " + chID)
		}
	case "2":
		var chID string
		fmt.Println("Введите ID чата для начала общения или введите return для выхода в предыдущее меню.")
		_, err := fmt.Scanln(&chID)
		if err != nil {
			app.logger.Error().Err(err).Msg("failed to scan chat ID")
			color.Red("Ошибка ввода ID чата")
			return false
		}
		switch chID {
		case "return":
		default:
			app.chatID = chID
			app.curentState = inChat
		}
	case "exit":
		return false
	default:
		color.Red("Такого действия нет")
	}

	return true
}

func (app *app) requestNewChat(userID string) (string, error) {
	conn, closeFn := app.connectFn()
	defer closeFn()

	var req domain.Request
	req.Type = domain.ReqTypeNewChat

	var chreq domain.NewChatRequest
	chreq.UserIDs = []domain.ID{domain.ID(userID)}
	data, err := json.Marshal(chreq)
	if err != nil {
		return "", err
	}
	req.Data = data

	if err := conn.WriteJSON(req); err != nil {
		return "", err
	}

	var resp domain.Delivery
	if err := conn.ReadJSON(&resp); err != nil {
		return "", fmt.Errorf("read json from conn error: %w", err)
	}
	chatId, ok := resp.Data.(string)
	if !ok {
		return "", fmt.Errorf("server error: not string")
	}

	return chatId, nil
}

func (app *app) chat(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)

	conn, closeFn := app.connectFn()
	defer closeFn()

	go func() {
		for {
			var msg domain.Delivery
			if err := conn.ReadJSON(&msg); err != nil {
				// conn will be closed in defer and returned error here
				var clErr *websocket.CloseError
				if errors.As(err, &clErr) {
					if clErr.Code != websocket.CloseNormalClosure {
						app.logger.Error().Err(clErr).Msg("read json from conn: close conn")
					}
				} else {
					app.logger.Error().Err(err).Msg("read json from conn")
				}
				cancel()
				return
			}
			if msg.Type == domain.DeliveryTypeNewMsg {
				switch v := msg.Data.(type) {
				case *domain.MessageChatDelivery:
					if v.ChID != domain.ID(app.chatID) {
						continue
					}
					color.Blue(v.Message.Body)
				case *domain.ErrorResponse:
					app.logger.Error().Str("error", v.Error).Msg("server error")
					color.Red("Ошибка: " + v.Error)
				}
			}
		}
	}()

	req := domain.Request{
		Type: domain.ReqTypeNewMsg,
	}

	mchreq := domain.MessageChatRequest{
		ChID: domain.ID(app.chatID),
		Type: domain.MsgTypeAdd,
	}

	go func() {
		sc := bufio.NewScanner(os.Stdin)
		sc.Split(bufio.ScanLines)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				if !sc.Scan() {
					color.Red("Ошибка ввода данных")
					app.logger.Error().Err(sc.Err()).Msg("read from console")
					continue
				}
				if sc.Text() == returnText {
					cancel()
					return
				}
				mchreq.Msg = sc.Text()

				data, err := json.Marshal(mchreq)
				if err != nil {
					color.Red("Ошибка преобразования данных")
					app.logger.Error().Err(err).Msg("json marshal error")
					cancel()
					return
				}
				req.Data = data

				if err := conn.WriteJSON(req); err != nil {
					color.Red("Ошибка отправки данных")
					app.logger.Error().Err(err).Msg("conn write")
					cancel()
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}()

	<-ctx.Done()
}
