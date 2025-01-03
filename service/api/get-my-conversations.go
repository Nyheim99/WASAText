package api

import (
	"encoding/json"
	"net/http"
	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

func (rt *_router) getUserConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Retrieve the authenticated userId from context
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}

	userId := reqCtx.UserID

	// Fetch conversations for the authenticated user
	conversations, err := rt.db.GetUserConversations(userId)
	if err != nil {
		http.Error(w, "Failed to fetch conversations", http.StatusInternalServerError)
		return
	}

	// Respond with the conversations as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}