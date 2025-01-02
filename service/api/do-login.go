package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	type loginRequest struct {
		Name string `json:"name"`
	}
	type loginResponse struct {
		Identifier string `json:"identifier"`
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Logic to create or retrieve user
	identifier := "generated-user-id" // Replace with actual user logic

	resp := loginResponse{Identifier: identifier}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}