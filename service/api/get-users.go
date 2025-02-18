package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Fetch users from database
	users, err := rt.db.GetUsers()
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	//Validate that at least 1 user exists
	if len(users) == 0 {
		http.Error(w, "No users found", http.StatusNotFound)
		return
	}

	//Set cap to 100 users
	if len(users) > 100 {
		users = users[:100]
	}

	//Return the list of users
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
