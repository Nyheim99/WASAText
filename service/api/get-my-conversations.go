package api

import (
	"encoding/json"
	"net/http"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Get the user ID
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID := reqCtx.UserID

	//Get conversations from Database
	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	//Set a cap at 50 conversations
	if len(conversations) > 50 {
		conversations = conversations[:50]
	}

	w.Header().Set("Content-Type", "application/json")

	//If the user has no conversations, return an empty string array
	if len(conversations) == 0 {
		err = json.NewEncoder(w).Encode([]string{})
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	//Otherwise return the conversation list
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(conversations)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
