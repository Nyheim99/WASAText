package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

type SendMessageResponse struct {
	MessageID   int64  `json:"message_id"`
	MessageType string `json:"message_type"`
	Content     string `json:"content"`
	SenderID    int64  `json:"sender_id"`
	SenderName  string `json:"sender_name"`
	Timestamp   string `json:"timestamp"`
}

func (rt *_router) sendMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationIDStr := ps.ByName("conversationID")
	conversationID, err := strconv.ParseInt(conversationIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	senderID := reqCtx.UserID

	message := r.FormValue("message")
	var textContent *string
	if message != "" {
		textContent = &message
	}

	var photoData *[]byte
	var photoMimeType *string
	file, handler, err := r.FormFile("photo")
	if err == nil {
		defer file.Close()

		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, file); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		data := buf.Bytes()
		photoData = &data

		mimeType := strings.ToLower(filepath.Ext(handler.Filename))
		if mimeType == ".jpg" || mimeType == ".jpeg" {
			mimeType = "image/jpeg"
		} else if mimeType == ".png" {
			mimeType = "image/png"
		} else {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		photoMimeType = &mimeType
	}

	if textContent == nil && photoData == nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	messageID, err := rt.db.SendMessage(conversationID, senderID, textContent, photoData, photoMimeType)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sender, err := rt.db.GetUser(senderID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if sender == nil {
		http.Error(w, "Sender not found", http.StatusNotFound)
		return
	}

	messageType := "text"
	var content string
	if textContent != nil {
		content = *textContent

		if len(*textContent) < 1 || len(*textContent) > 1000 {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if match, _ := regexp.MatchString(`^[a-zA-Z0-9À-ÿ.,!?()\-\"' ]+$`, *textContent); !match {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}
	if photoData != nil {
		messageType = "photo"
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(SendMessageResponse{
		MessageID:   messageID,
		MessageType: messageType,
		Content:     content,
		SenderID:    senderID,
		SenderName:  sender.Username,
		Timestamp:   timestamp,
	})
}
