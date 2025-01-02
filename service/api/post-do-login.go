package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
)

type loginRequest struct {
	Username string `json:"username"`
}

type loginResponse struct {
	Identifier int64 `json:"identifier"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate username length
	if len(request.Username) < 3 || len(request.Username) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}

	// Check if the user exists in the database
	identifier, err := rt.db.GetUserByUsername(request.Username)
	if err != nil {
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }

	// If the user doesn't exist, create a new one
	if identifier == 0 {
		identifier, err = rt.db.CreateUser(request.Username)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	// Respond with the user's identifier
	response := loginResponse{Identifier: identifier}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}