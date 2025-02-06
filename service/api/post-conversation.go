package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	conversationType := r.FormValue("conversation_type")

	if conversationType == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID := reqCtx.UserID

	var conversationID int64
	var err error

	if conversationType == "private" {
		recipientIDStr := r.FormValue("recipientID")

		recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		conversationID, err = rt.db.CreatePrivateConversation(userID, recipientID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
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

		conversationID, err = rt.db.CreateGroupConversation(userID, groupName, "", participantIDs)
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
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	message := r.FormValue("message")

	if message == ""  {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	_, err = rt.db.SendMessage(conversationID, userID, &message, nil, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
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
