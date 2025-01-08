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
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Authorization header missing", http.StatusUnauthorized)
            return
        }

        if !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        userId, err := strconv.ParseInt(token, 10, 64)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        user, err := rt.db.GetUser(userId)
        if err != nil || user == nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        reqUUID, err := uuid.NewV4()
        if err != nil {
            http.Error(w, "Failed to generate request UUID", http.StatusInternalServerError)
            return
        }

        logger := logrus.WithFields(logrus.Fields{
            "reqUUID": reqUUID,
            "userId":  userId,
        })

        reqCtx := &reqcontext.RequestContext{
            ReqUUID: reqUUID,
            Logger:  logger,
            UserID:  userId,
        }

        ctx := context.WithValue(r.Context(), "reqCtx", reqCtx)
        r = r.WithContext(ctx)

        next(w, r, ps)
    }
}