package api

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (rt *_router) getUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	users := []map[string]interface{}{
		{"id": 1, "username": "Alice", "photo_url": "https://example.com/alice.jpg"},
		{"id": 2, "username": "Bob", "photo_url": "https://example.com/bob.jpg"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}