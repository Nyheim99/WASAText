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

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationIDStr := ps.ByName("conversationID")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	UserID := reqCtx.UserID

	var request AddToGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if len(request.Participants) == 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conversation.ConversationType != "group" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	isMember := false
	for _, participant := range conversation.Participants {
		if participant.ID == UserID {
			isMember = true
			break
		}
	}

	if !isMember {
		http.Error(w, "User is not a member of the conversation", http.StatusForbidden)
		return
	}

	err = rt.db.AddToGroup(conversationID, request.Participants)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
