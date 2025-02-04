package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

// DeleteMessageResponse defines the API response structure
type DeleteMessageResponse struct {
	MessageID      int64  `json:"message_id"`
	ConversationID int64  `json:"conversation_id"`
	Status        string `json:"status"`
}

// deleteMessage handles soft deletion of messages
func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract conversation ID from URL
	conversationID, err := strconv.ParseInt(ps.ByName("conversationID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Extract message ID from URL
	messageID, err := strconv.ParseInt(ps.ByName("messageID"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	// Retrieve the user ID from request context
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}
	userID := reqCtx.UserID

	// Attempt to soft delete the message
	err = rt.db.DeleteMessage(conversationID, messageID, userID)
	if err != nil {
		if err.Error() == "message not found or already deleted" {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, fmt.Sprintf("Failed to delete message: %s", err), http.StatusInternalServerError)
		}
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(DeleteMessageResponse{
		MessageID:      messageID,
		ConversationID: conversationID,
		Status:         "deleted",
	})
}
