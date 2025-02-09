package api

import (
	"net/http"
	"strconv"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) markMessagesAsRead(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	userID := reqCtx.UserID

	err = rt.db.MarkMessagesAsRead(conversationID, userID)
	if err != nil {
		http.Error(w, "Failed to mark messages as read", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
