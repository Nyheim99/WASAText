package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
)

type loginRequest struct {
	Username string `json:"username"`
}

type loginResponse struct {
	Identifier int64 `json:"identifier"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Validate request
	var request loginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(request.Username) < 3 || len(request.Username) > 16 {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9]*$`)
	if !validUsername.MatchString(request.Username) {
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	//Check if username exists
	identifier, err := rt.db.GetUserByUsername(request.Username)
	if err != nil {
		http.Error(w, "Failed to retrieve user from database", http.StatusInternalServerError)
		return
	}

	//If it does not, create a new user
	if identifier == 0 {
		identifier, err = rt.db.CreateUser(request.Username)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	}

	//Return the identifier
	response := loginResponse{Identifier: identifier}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Authorization", fmt.Sprintf("Bearer %d", identifier))
	w.WriteHeader(http.StatusCreated)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
