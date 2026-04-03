package support

import (
	"encoding/json"
	"fmt"
	"kroncl-server/internal/config"
	"kroncl-server/internal/core"
	"kroncl-server/internal/tenant/logs"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const (
	EVENT_MESSAGE_PING          = "ping"
	EVENT_MESSAGE_PONG          = "pong"
	EVENT_MESSAGES_NEW_MESSAGE  = "new_message"
	EVENT_MESSAGES_MESSAGE_READ = "message_read"
	EVENT_MESSAGES_MARK_READ    = "mark_read"
)

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageReadPayload struct {
	MessageID string `json:"message_id"`
	Read      bool   `json:"read"`
}

type wsClient struct {
	conn      *websocket.Conn
	accountID string
}

type wsHub struct {
	clients map[string]map[*websocket.Conn]string // ticketID -> conn -> accountID
	mu      sync.RWMutex
}

var hub = &wsHub{
	clients: make(map[string]map[*websocket.Conn]string),
}

func (h *wsHub) addClient(ticketID string, conn *websocket.Conn, accountID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[ticketID] == nil {
		h.clients[ticketID] = make(map[*websocket.Conn]string)
	}
	h.clients[ticketID][conn] = accountID
}

func (h *wsHub) removeClient(ticketID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[ticketID] != nil {
		delete(h.clients[ticketID], conn)
		if len(h.clients[ticketID]) == 0 {
			delete(h.clients, ticketID)
		}
	}
}

func (h *wsHub) broadcastToTicket(ticketID string, message WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.clients[ticketID] == nil {
		return
	}

	for conn := range h.clients[ticketID] {
		if err := conn.WriteJSON(message); err != nil {
			// Ошибка записи — клиент, вероятно, отключился
			// Очистка произойдёт при закрытии соединения
			_ = err
		}
	}
}

// BroadcastNewMessage уведомляет всех о новом сообщении
func BroadcastNewMessage(ticketID string, message *Message) {
	hub.broadcastToTicket(ticketID, WebSocketMessage{
		Type:    EVENT_MESSAGES_NEW_MESSAGE,
		Payload: message,
	})
}

// BroadcastMessageRead уведомляет о прочтении сообщения
func BroadcastMessageRead(ticketID, messageID string, read bool) {
	hub.broadcastToTicket(ticketID, WebSocketMessage{
		Type: EVENT_MESSAGES_MESSAGE_READ,
		Payload: MessageReadPayload{
			MessageID: messageID,
			Read:      read,
		},
	})
}

// MessagesWebSocket обрабатывает WebSocket соединение
func (h *Handlers) MessagesWebSocket(w http.ResponseWriter, r *http.Request) {
	accountID, ok := core.GetUserIDFromContext(r.Context())
	if !ok {
		core.SendError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	companyID := chi.URLParam(r, "id")
	if companyID == "" {
		core.SendValidationError(w, "Company ID required")
		return
	}

	ticketID := chi.URLParam(r, "ticketId")
	if ticketID == "" {
		core.SendValidationError(w, "Ticket ID required")
		return
	}

	// Проверяем доступ к тикету
	if err := h.service.CheckTicketAccess(r.Context(), companyID, ticketID); err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", err.Error()),
			logs.WithMetadata("ticket_id", ticketID),
		)
		core.SendNotFound(w, "Ticket not found")
		return
	}

	// Upgrade connection
	conn, err := config.WebSocketUpgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
			logs.WithStatus(logs.LogStatusError),
			logs.WithUserAgent(r.UserAgent()),
			logs.WithMetadata("error", fmt.Sprintf("WebSocket upgrade failed: %v", err)),
		)
		core.SendInternalError(w, "WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	// Регистрируем клиента
	hub.addClient(ticketID, conn, accountID)
	defer hub.removeClient(ticketID, conn)

	h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
		logs.WithStatus(logs.LogStatusSuccess),
		logs.WithUserAgent(r.UserAgent()),
		logs.WithMetadata("ticket_id", ticketID),
		logs.WithMetadata("action", "websocket_connected"),
	)

	// Отправляем приветствие
	conn.WriteJSON(WebSocketMessage{
		Type:    "connected",
		Payload: map[string]string{"ticket_id": ticketID},
	})

	// Слушаем сообщения от клиента
	for {
		var msg WebSocketMessage
		if err := conn.ReadJSON(&msg); err != nil {
			break
		}

		switch msg.Type {
		case EVENT_MESSAGE_PING:
			conn.WriteJSON(WebSocketMessage{Type: EVENT_MESSAGE_PONG})
		case EVENT_MESSAGES_MARK_READ:
			payloadBytes, _ := json.Marshal(msg.Payload)
			var readPayload MessageReadPayload
			if err := json.Unmarshal(payloadBytes, &readPayload); err == nil {
				go func() {
					if err := h.service.UpdateMessageReadStatus(r.Context(), readPayload.MessageID, readPayload.Read); err != nil {
						h.logsService.Log(r.Context(), config.PERMISSION_SUPPORT_TICKETS, accountID,
							logs.WithStatus(logs.LogStatusError),
							logs.WithUserAgent(r.UserAgent()),
							logs.WithMetadata("error", err.Error()),
							logs.WithMetadata("message_id", readPayload.MessageID),
						)
					}
				}()
			}
		}
	}
}
