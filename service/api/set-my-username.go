package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

type setMyUsernameRequest struct {
	Username string `json:"username"`
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}

	userId := reqCtx.UserID

	var req setMyUsernameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Invalid username length", http.StatusBadRequest)
		return
	}
	if !regexp.MustCompile("^[a-zA-Z0-9]*$").MatchString(req.Username) {
		http.Error(w, "Username must be alphanumeric", http.StatusBadRequest)
		return
	}

	exists, err := rt.db.DoesUsernameExist(req.Username)
	if err != nil {
		http.Error(w, "Failed to check username", http.StatusInternalServerError)
		fmt.Println("Database error:", err)
		return
	}
	if exists {
		http.Error(w, "Username already in use", http.StatusConflict)
		return
	}

	err = rt.db.SetMyUserName(userId, req.Username)
	if err != nil {
		http.Error(w, "Failed to update username", http.StatusInternalServerError)
		fmt.Println("Database error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
			"message": "Username updated successfully",
			"username": req.Username,
	})
}
