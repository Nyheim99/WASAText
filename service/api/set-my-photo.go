package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the user from the request context
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}

	userID := reqCtx.UserID

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	if !allowedExtensions[fileExt] {
		http.Error(w, "Invalid file type. Only JPG and PNG are allowed", http.StatusBadRequest)
		return
	}

	// Define file save path
	fileName := fmt.Sprintf("user_%d%s", userID, fileExt)
	savePath := filepath.Join("service/profile_pictures", fileName)

	// Ensure directory exists
	err = os.MkdirAll("service/profile_pictures", os.ModePerm)
	if err != nil {
		http.Error(w, "Server error: unable to create directory", http.StatusInternalServerError)
		return
	}

	// Save the new file
	out, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Server error: unable to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Server error: unable to save file", http.StatusInternalServerError)
		return
	}

	// Update the photo URL in the database
	photoURL := fmt.Sprintf("/service/profile_pictures/%s", fileName)
	err = rt.db.SetMyPhoto(userID, photoURL)
	if err != nil {
		http.Error(w, "Failed to update profile picture", http.StatusInternalServerError)
		fmt.Println("Database error:", err)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Profile picture uploaded successfully",
		"photo_url": photoURL,
	})
}
