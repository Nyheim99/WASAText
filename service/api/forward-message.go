package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type ForwardMessageResponse struct {
	MessageID   int64  `json:"message_id"`
	MessageType string `json:"message_type"`
	OriginalID  int64  `json:"original_message_id"`
	SenderID    int64  `json:"sender_id"`
	Timestamp   string `json:"timestamp"`
}

func (rt *_router) forwardMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationIDStr := ps.ByName("conversationID")
	originalMessageIDStr := ps.ByName("messageID")

	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	originalMessageID, err := strconv.ParseInt(originalMessageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	senderID := reqCtx.UserID

	messageID, err := rt.db.ForwardMessage(conversationID, senderID, originalMessageID)
	if err != nil {
		http.Error(w, "Failed to forward message", http.StatusInternalServerError)
		return
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ForwardMessageResponse{
		MessageID:   messageID,
		MessageType: "forwarded",
		OriginalID:  originalMessageID,
		SenderID:    senderID,
		Timestamp:   timestamp,
	})
}
