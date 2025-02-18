package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type commentMessageRequest struct {
	Emoticon string `json:"emoticon"`
}

//Comment a message
func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Get message ID
	messageIDStr := ps.ByName("messageID")
	messageID, err := strconv.ParseInt(messageIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Get user ID
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userID := reqCtx.UserID

	//Validate request
	if userID <= 0 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var req commentMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if req.Emoticon == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Comment the message
	err = rt.db.CommentMessage(messageID, userID, req.Emoticon)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
