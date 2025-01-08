package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/Nyheim99/WASAText/service/api/reqcontext"
)

// ConversationPreview represents a preview of a conversation for the API response.
type ConversationPreview struct {
	ConversationID      int64  `json:"conversation_id"`
	ConversationType    string `json:"conversation_type"`
	DisplayName         string `json:"display_name"`
	DisplayPhotoURL     string `json:"display_photo_url"`
	LastMessageContent  string `json:"last_message_content,omitempty"`
	LastMessagePhotoURL string `json:"last_message_photo_url,omitempty"`
	LastMessageTimestamp string `json:"last_message_timestamp"`
}

// ConversationListResponse represents the API response for the list of conversations.
type ConversationListResponse struct {
	Conversations []ConversationPreview `json:"conversations"`
}

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Retrieve the request context to access the user ID
	reqCtx, ok := r.Context().Value("reqCtx").(*reqcontext.RequestContext)
	if !ok || reqCtx == nil {
		http.Error(w, "Request context missing", http.StatusInternalServerError)
		return
	}
	userID := reqCtx.UserID

	if userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Call the database function to get conversations
	dbConversations, err := rt.db.GetMyConversations(userID)
	if err != nil {
		http.Error(w, "Failed to retrieve conversations", http.StatusInternalServerError)
		fmt.Printf("Database error: %v\n", err)
		return
	}

	// Map database results to API struct
	apiConversations := make([]ConversationPreview, len(dbConversations))
	for i, conv := range dbConversations {
		apiConversations[i] = ConversationPreview{
			ConversationID:      conv.ConversationID,
			ConversationType:    conv.ConversationType,
			DisplayName:         conv.DisplayName,
			DisplayPhotoURL:     conv.DisplayPhotoURL,
			LastMessageContent:  conv.LastMessageContent,
			LastMessagePhotoURL: conv.LastMessagePhotoURL,
			LastMessageTimestamp: conv.LastMessageTimestamp,
		}
	}

	// Prepare the response
	response := ConversationListResponse{
		Conversations: apiConversations,
	}

	// Encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		fmt.Printf("JSON encoding error: %v\n", err)
		return
	}
}
