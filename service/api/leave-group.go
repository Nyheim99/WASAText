package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

// LeaveGroupResponse defines the response format for leaving a group
type LeaveGroupResponse struct {
	ConversationID int64  `json:"conversation_id"`
	Status         string `json:"status"`
}

// leaveGroup allows a user to leave a group conversation.
func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Extract conversation ID from URL parameters
	conversationIDStr := ps.ByName("conversationID")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Retrieve request context
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}
	currentUserID := reqCtx.UserID

	// Check if the conversation exists and is a group
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conversation.ConversationType != "group" {
		http.Error(w, "Cannot leave a private conversation", http.StatusBadRequest)
		return
	}

	// Verify that the user is part of the group
	isMember := false
	for _, participant := range conversation.Participants {
		if participant.ID == currentUserID {
			isMember = true
			break
		}
	}
	if !isMember {
		http.Error(w, "User is not a member of this group", http.StatusForbidden)
		return
	}

	// Call the database function to remove the user from the group
	err = rt.db.LeaveGroup(conversationID, currentUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LeaveGroupResponse{
		ConversationID: conversationID,
		Status:         "User successfully left the group",
	})
}
