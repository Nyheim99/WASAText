package api

import (
    "encoding/json"
    "fmt"
    "net/http"

    "github.com/julienschmidt/httprouter"
    "your_project_path/service/api/reqcontext"
)

type updateUsernameRequest struct {
    Username string `json:"username"`
}

type updateUsernameResponse struct {
    UserID   int64  `json:"userId"`
    Username string `json:"username"`
}

func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
    // Retrieve the authenticated userId from the context
    reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
    if !ok || reqCtx == nil {
        http.Error(w, "Request context missing", http.StatusInternalServerError)
        return
    }

    userId := reqCtx.UserID

    // Parse the request body
    var req updateUsernameRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate the username
    if len(req.Username) < 3 || len(req.Username) > 16 {
        http.Error(w, "Username must be between 3 and 16 characters", http.StatusBadRequest)
        return
    }

    // Update the username in the database
    err := rt.db.UpdateUserName(userId, req.Username)
    if err != nil {
        if err.Error() == "username already in use" {
            http.Error(w, "Username already in use", http.StatusConflict)
        } else {
            http.Error(w, "Failed to update username", http.StatusInternalServerError)
        }
        return
    }

    // Respond with the updated user data
    res := updateUsernameResponse{
        UserID:   userId,
        Username: req.Username,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}