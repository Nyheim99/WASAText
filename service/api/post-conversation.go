package api

import (
	"encoding/json"
	"net/http"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"mime/multipart"
	"bytes"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

type NewConversationResponse struct {
	ConversationID int64 `json:"conversation_id"`
	MessageID      int64 `json:"message_id"`
}

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse multipart form-data
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	// Extract common form fields
	conversationType := r.FormValue("conversation_type")
	message := r.FormValue("message")

	if conversationType == "" {
		http.Error(w, "conversation_type is required", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}
	currentUserID := reqCtx.UserID

	var conversationID int64
	var err error

	if conversationType == "private" {
		username := r.FormValue("username")
		if username == "" {
			http.Error(w, "username is required for private conversation", http.StatusBadRequest)
			return
		}

		// Get recipient user ID
		recipientID, err := rt.db.GetUserByUsername(username)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		if recipientID == 0 {
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}

		// Create or fetch private conversation
		conversationID, err = rt.db.GetOrCreatePrivateConversation(currentUserID, recipientID)
		if err != nil {
			http.Error(w, "Failed to create private conversation", http.StatusInternalServerError)
			return
		}
	} else if conversationType == "group" {
		groupName := r.FormValue("group_name")
		participants := r.FormValue("participants")

		if groupName == "" || participants == "" {
			http.Error(w, "group_name and participants are required for group conversation", http.StatusBadRequest)
			return
		}

		var participantIDs []int64
		err := json.Unmarshal([]byte(participants), &participantIDs)
		if err != nil {
			http.Error(w, "Invalid participants format", http.StatusBadRequest)
			return
		}

		conversationID, err = rt.db.CreateGroupConversation(currentUserID, groupName, "", participantIDs)
		if err != nil {
			http.Error(w, "Failed to create group conversation", http.StatusInternalServerError)
			return
		}

		// Handle optional group photo
		var photoURL string
		file, handler, err := r.FormFile("group_photo")
		if err == nil {
			defer file.Close()
			photoURL, err = rt.saveUploadedFile(file, handler, conversationID)
			if err != nil {
				http.Error(w, "Failed to save group photo", http.StatusInternalServerError)
				return
			}
			err = rt.db.SetGroupPhoto(conversationID, photoURL)
			if err != nil {
				http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
				return
			}
		}
	} else {
		http.Error(w, "Unsupported conversation type", http.StatusBadRequest)
		return
	}

	// Handle Message (Text or Photo)
	var messageID int64
	var textContent *string
	var photoData *[]byte
	var photoMimeType *string

	if message != "" {
		textContent = &message
	}

	// Handle optional photo message
	file, handler, err := r.FormFile("message_photo")
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

	// Ensure at least one of text or photo is present
	if textContent == nil && photoData == nil {
		http.Error(w, "Either message text or a photo is required", http.StatusBadRequest)
		return
	}

	// Store message in database
	messageID, err = rt.db.SendMessage(conversationID, currentUserID, textContent, photoData, photoMimeType)
	if err != nil {
		http.Error(w, "Failed to add message", http.StatusInternalServerError)
		return
	}

	// Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(NewConversationResponse{
		ConversationID: conversationID,
		MessageID:      messageID,
	})
}

// Helper function to save uploaded group photos
func (rt *_router) saveUploadedFile(file io.Reader, handler *multipart.FileHeader, conversationID int64) (string, error) {
	// Extract and validate file extension
	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExtensions[fileExt] {
		return "", fmt.Errorf("invalid file type. Only JPG and PNG are allowed")
	}

	// Define file name as "group_<conversationID>.<ext>"
	fileName := fmt.Sprintf("group_%d%s", conversationID, fileExt)

	// Define file path
	savePath := filepath.Join("service/photos/groups", fileName)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create directory: %w", err)
	}

	// Save file to disk
	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("unable to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("unable to save file: %w", err)
	}

	// Return correct photo URL
	return fmt.Sprintf("/service/photos/groups/%s", fileName), nil
}