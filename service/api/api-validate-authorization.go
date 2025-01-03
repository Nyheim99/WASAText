package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/sirupsen/logrus"
	"github.com/gofrs/uuid"
)

func (rt *_router) validateAuthorization(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Extract the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		// Extract the token (userId)
		token := strings.TrimPrefix(authHeader, "Bearer ")
		userId, err := strconv.ParseInt(token, 10, 64)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Validate the userId in the database
		exists, err := rt.db.DoesUserExist(userId)
		if err != nil || !exists {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Generate a unique request UUID
		reqUUID, err := uuid.NewV4()
		if err != nil {
			http.Error(w, "Failed to generate request UUID", http.StatusInternalServerError)
			return
		}

		// Create a logger for the request
		logger := logrus.WithFields(logrus.Fields{
			"reqUUID": reqUUID,
			"userId":  userId,
		})

		// Populate the RequestContext
		reqCtx := &reqcontext.RequestContext{
			ReqUUID: reqUUID,
			Logger:  logger,
			UserID:  userId,
		}

		// Add RequestContext to the request context
		ctx := context.WithValue(r.Context(), "reqCtx", reqCtx)
		r = r.WithContext(ctx)

		// Call the next handler
		next(w, r, ps)
	}
}
