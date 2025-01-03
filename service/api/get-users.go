package api

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
	"encoding/json"
	"strconv"
	"fmt"
)

func (rt *_router) getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query()
	var conversationID *int64

	// Parse conversationId if present
	if idStr := query.Get("conversationId"); idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid conversationId", http.StatusBadRequest)
			fmt.Println("Error parsing conversationId:", err)
			return
		}
		conversationID = &id
	}

	// Fetch users from the database
	users, err := rt.db.GetUsers(conversationID)
	if err != nil {
		http.Error(w, "Failed to fetch users", http.StatusInternalServerError)
		fmt.Println("Database error in getUsers:", err)
		return
	}

	// Respond with the user data as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		fmt.Println("JSON encoding error in getUsers:", err)
	}
}