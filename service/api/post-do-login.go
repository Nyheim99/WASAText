package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"fmt"
	"regexp"
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

    if len(request.Username) < 3 || len(request.Username) > 16 {
        http.Error(w, "Invalid username length", http.StatusBadRequest)
        return
    }
    validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
    if !validUsername.MatchString(request.Username) {
        http.Error(w, "Invalid username format", http.StatusBadRequest)
        return
    }

    identifier, err := rt.db.GetUserByUsername(request.Username)
    if err != nil {
        http.Error(w, "Failed to retrieve user from database", http.StatusInternalServerError)
        return
    }

    if identifier == 0 {
        identifier, err = rt.db.CreateUser(request.Username)
        if err != nil {
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }
    }

    response := loginResponse{Identifier: identifier}
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Authorization", fmt.Sprintf("Bearer %d", identifier))
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}