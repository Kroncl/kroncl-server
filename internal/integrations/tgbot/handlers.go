package tgbot

import (
	"encoding/json"
	"net/http"

	"github.com/mymmrac/telego"
)

type Handlers struct {
	service *Service
}

func NewHandlers(service *Service) *Handlers {
	return &Handlers{service: service}
}

func (h *Handlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	secret := r.Header.Get("X-Telegram-Bot-Api-Secret-Token")
	if secret != h.service.cfg.WebhookSecret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var update telego.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if update.Message != nil {
		h.service.bot.SendMessage(r.Context(), &telego.SendMessageParams{
			ChatID: telego.ChatID{ID: update.Message.Chat.ID},
			Text:   "все заебись братан",
		})
	}

	w.WriteHeader(http.StatusOK)
}
