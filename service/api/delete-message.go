package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

type DeleteMessageResponse struct {
	MessageID      int64  `json:"message_id"`
	ConversationID int64  `json:"conversation_id"`
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationID, err := strconv.ParseInt(ps.ByName("conversationID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	messageID, err := strconv.ParseInt(ps.ByName("messageID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userID := reqCtx.UserID

	err = rt.db.DeleteMessage(conversationID, messageID, userID)
	if err != nil {
		if err.Error() == "message not found or already deleted" {
			http.Error(w, "Message not found or already deleted", http.StatusNotFound)
		} else {
			http.Error(w, "Invalid request", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteMessageResponse{
		MessageID:      messageID,
		ConversationID: conversationID,
	})
}
