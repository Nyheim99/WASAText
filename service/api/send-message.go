package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"strconv"
	"time"
	
	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
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
        http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
        return
    }

    // Parse multipart form-data
    if err := r.ParseMultipartForm(10 << 20); err != nil {
        http.Error(w, "Failed to parse form data", http.StatusBadRequest)
        return
    }

    // Retrieve request context
    reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
    if !ok || reqCtx == nil {
        http.Error(w, "Request context missing", http.StatusInternalServerError)
        return
    }
    senderID := reqCtx.UserID

    // Extract message text (if provided)
    message := r.FormValue("message")
    var textContent *string
    if message != "" {
        textContent = &message
    }

    // Handle photo message (if provided)
    var photoData *[]byte
    var photoMimeType *string
    file, handler, err := r.FormFile("photo")
    if err == nil {
        defer file.Close()

        // Read file into byte slice
        buf := new(bytes.Buffer)
        if _, err := io.Copy(buf, file); err != nil {
            http.Error(w, "Failed to read uploaded image", http.StatusInternalServerError)
            return
        }
        data := buf.Bytes()
        photoData = &data

        // Extract MIME type from file extension
        mimeType := strings.ToLower(filepath.Ext(handler.Filename))
        if mimeType == ".jpg" || mimeType == ".jpeg" {
            mimeType = "image/jpeg"
        } else if mimeType == ".png" {
            mimeType = "image/png"
        } else {
            http.Error(w, "Invalid image type. Only JPG and PNG are allowed.", http.StatusBadRequest)
            return
        }
        photoMimeType = &mimeType
    }

    // Ensure at least one of text or photo is provided
    if textContent == nil && photoData == nil {
        http.Error(w, "Either message text or a photo is required", http.StatusBadRequest)
        return
    }

    // Send the message
    messageID, err := rt.db.SendMessage(conversationID, senderID, textContent, photoData, photoMimeType)
    if err != nil {
        http.Error(w, "Failed to send message", http.StatusInternalServerError)
        return
    }

    // Fetch sender's information (using the senderID)
    sender, err := rt.db.GetUser(senderID)
    if err != nil {
        http.Error(w, "Failed to fetch sender information", http.StatusInternalServerError)
        return
    }
    if sender == nil {
        http.Error(w, "Sender not found", http.StatusNotFound)
        return
    }

    // Prepare the message object
    messageType := "text"
    var content string
    if textContent != nil {
        content = *textContent
    }
    if photoData != nil {
        messageType = "photo"
    }

    // Format the timestamp
    timestamp := time.Now().UTC().Format(time.RFC3339)

    // Respond with the complete message object
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
