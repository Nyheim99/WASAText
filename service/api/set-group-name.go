package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type setGroupNameRequest struct {
	Name string `json:"name"`
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationID := ps.ByName("conversationID")
	if conversationID == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var req setGroupNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Name) < 3 || len(req.Name) > 20 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if match, _ := regexp.MatchString(`^[a-zA-Z0-9 ]*$`, req.Name); !match {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = rt.db.SetGroupName(convID, req.Name)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(req.Name)
}
