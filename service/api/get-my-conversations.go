package api

import (
	"encoding/json"
	"net/http"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID := reqCtx.UserID

	conversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(conversations) > 50 {
		conversations = conversations[:50]
	}

	w.Header().Set("Content-Type", "application/json")

	if len(conversations) == 0 {
		err = json.NewEncoder(w).Encode([]string{})
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(conversations)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
