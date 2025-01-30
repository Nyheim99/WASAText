package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

type AddToGroupRequest struct {
	Participants []int64 `json:"participants"`
}

type AddToGroupResponse struct {
	ConversationID int64   `json:"conversation_id"`
	AddedUsers     []int64 `json:"added_users"`
}

// addToGroup handles adding new members to an existing group conversation.
func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationIDStr := ps.ByName("conversationID")
conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
if err != nil {
	http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
	return
}

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

	// Decode request body
	var request AddToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(request.Participants) == 0 {
		http.Error(w, "At least one participant must be provided", http.StatusBadRequest)
		return
	}

	// Check if the conversation is a group
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conversation.ConversationType != "group" {
		http.Error(w, "Cannot add members to a private conversation", http.StatusBadRequest)
		return
	}

	// Verify that the current user is part of the group
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

	err = rt.db.AddToGroup(conversationID, request.Participants)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AddToGroupResponse{
		ConversationID: conversationID,
		AddedUsers:     request.Participants,
	})
}
