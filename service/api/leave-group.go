package api

import (
	"net/http"
	"strconv"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Get conversation ID
	conversationIDStr := ps.ByName("conversationID")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	//Get user ID
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	UserID := reqCtx.UserID

	//Fetch vonersation from database
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	if conversation.ConversationType != "group" {
		http.Error(w, "Cannot leave a private conversation", http.StatusBadRequest)
		return
	}

	//Check that the user is member of the group
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

	//Leave the group
	err = rt.db.LeaveGroup(conversationID, UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
