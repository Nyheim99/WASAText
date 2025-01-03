package api

import (
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
    "net/http"
    "fmt"
)

// httpRouterHandler is the signature for functions that accepts a reqcontext.RequestContext in addition to those
// required by the httprouter package.
type httpRouterHandler func(http.ResponseWriter, *http.Request, httprouter.Params, reqcontext.RequestContext)

func (rt *_router) validateAuthorization(next httprouter.Handle) httprouter.Handle {
    return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
        // Extract Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        // Parse user identifier from the header
        var userID int64
        n, err := fmt.Sscanf(authHeader, "Bearer %d", &userID)
        if err != nil || n != 1 {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }

        // Validate the user exists in the database
        exists, err := rt.db.DoesUserExist(userID)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        if !exists {
            http.Error(w, "Invalid user identifier", http.StatusUnauthorized)
            return
        }

        // Pass control to the next handler
        next(w, r, ps)
    }
}
