package api

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Nyheim99/WASAText/service/api/reqcontext"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) createConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	//Validate request
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Get conversation type
	conversationType := r.FormValue("conversation_type")

	if conversationType == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//Get user ID
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID := reqCtx.UserID

	var conversationID int64
	var err error
	sendMessage := true


	//Check if conversation is private and if its created from a forwarded message
	if conversationType == "private" || conversationType == "new_user" {

		//get recipient ID
		recipientIDStr := r.FormValue("recipientID")

		recipientID, err := strconv.ParseInt(recipientIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		//Create the private conversation
		conversationID, err = rt.db.CreatePrivateConversation(userID, recipientID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//If the conversation is being created from a forwarded message, dont send a normal message
		if conversationType == "new_user" {
			sendMessage = false
		}

	} else if conversationType == "group" {

		//Get group name
		groupName := r.FormValue("group_name")

		//Validate group name
		if len(groupName) < 3 || len(groupName) > 20 {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if match, _ := regexp.MatchString(`^[a-zA-Z0-9 ]*$`, groupName); !match {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		//Get participants
		participantStrs := r.MultipartForm.Value["participants"]
		var participantIDs []int64

		for _, idStr := range participantStrs {
			id, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid participant ID format", http.StatusBadRequest)
				return
			}
			participantIDs = append(participantIDs, id)
		}

		//Validate participants
		if len(participantIDs) < 1 || len(participantIDs) > 50 {
			http.Error(w, "Invalid number of participants", http.StatusBadRequest)
			return
		}

		//Create the group conversation in the database
		conversationID, err = rt.db.CreateGroupConversation(userID, groupName, "", participantIDs)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		//Handle the group photo if attached
		if file, handler, err := r.FormFile("group_photo"); err == nil {
			defer file.Close()
			photoURL, err := rt.saveUploadedFile(file, handler, conversationID)
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if err = rt.db.SetGroupPhoto(conversationID, photoURL); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
	} else {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	//If the conversation is created normally, send the message
	if sendMessage {
		message := r.FormValue("message")

		//Validate the message
		if len(message) < 1 || len(message) > 1000 {
			http.Error(w, "Invalid message length", http.StatusBadRequest)
			return
		}

		if match, _ := regexp.MatchString(`^[a-zA-Z0-9À-ÿ.,!?()\-\"' ]+$`, message); !match {
			http.Error(w, "Invalid message format", http.StatusBadRequest)
			return
		}

		//Send the message in the database
		_, err = rt.db.SendMessage(conversationID, userID, &message, nil, nil, 0)
		if err != nil {
			http.Error(w, "Failed to send message", http.StatusInternalServerError)
			return
		}
	}

	//Return the conversation ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(map[string]int64{
		"conversation_id": conversationID,
	})
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}


//Helper function to save the optional group photo
func (rt *_router) saveUploadedFile(file io.Reader, handler *multipart.FileHeader, conversationID int64) (string, error) {
	fileExt := strings.ToLower(filepath.Ext(handler.Filename))
	allowedExtensions := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedExtensions[fileExt] {
		return "", fmt.Errorf("invalid file type. Only JPG and PNG are allowed")
	}

	fileName := fmt.Sprintf("group_%d%s", conversationID, fileExt)
	savePath := filepath.Join("service/photos/groups", fileName)

	if err := os.MkdirAll(filepath.Dir(savePath), os.ModePerm); err != nil {
		return "", fmt.Errorf("unable to create directory: %w", err)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("unable to create file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		return "", fmt.Errorf("unable to save file: %w", err)
	}

	return fmt.Sprintf("/service/photos/groups/%s", fileName), nil
}
