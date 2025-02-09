package api

import (
	"net/http"
	"strconv"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationIDStr := ps.ByName("conversationID")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	UserID := reqCtx.UserID

	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conversation.ConversationType != "group" {
		http.Error(w, "Cannot leave a private conversation", http.StatusBadRequest)
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

	err = rt.db.LeaveGroup(conversationID, UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
