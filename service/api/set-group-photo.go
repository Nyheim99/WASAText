package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	conversationID := ps.ByName("conversationID")
	if conversationID == "" {
		http.Error(w, "Conversation ID is required", http.StatusBadRequest)
		return
	}

	convID, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
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

	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	if !allowedExtensions[fileExt] {
		http.Error(w, "Invalid file type. Only JPG and PNG are allowed", http.StatusBadRequest)
		return
	}

	fileName := fmt.Sprintf("group_%d%s", convID, fileExt)
	savePath := filepath.Join("service/photos/groups", fileName)

	err = os.MkdirAll("service/photos/groups", os.ModePerm)
	if err != nil {
		http.Error(w, "Server error: unable to create directory", http.StatusInternalServerError)
		return
	}

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

	photoURL := fmt.Sprintf("/service/photos/groups/%s", fileName)
	err = rt.db.SetGroupPhoto(convID, photoURL)
	if err != nil {
		http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
		fmt.Println("Database error:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Group profile picture uploaded successfully",
		"photo_url": photoURL,
	})
}
