package database

import (
	"time"
)

type User struct {
	ID       int64  `json:"id"`         
	Username string `json:"username"`  
	PhotoURL string `json:"photo_url"`  
}

type Conversation struct {
	ID              int64  `json:"id"`                
	Name            string `json:"name,omitempty"`    
	ConversationType string `json:"conversation_type"` 
	PhotoURL        string `json:"photo_url"`        
	LastMessageID   int64  `json:"last_message_id"` 
}

type ConversationParticipant struct {
	ConversationID int64 `json:"conversation_id"` 
	UserID         int64 `json:"user_id"`       
}

type Message struct {
	ID                int64       `json:"id"`                  
	ConversationID    int64       `json:"conversation_id"`     
	SenderID          int64       `json:"sender_id"`          
	SenderUsername    string      `json:"sender_username"`
	Content           *string     `json:"content,omitempty"`
	PhotoData         *[]byte     `json:"photo_data,omitempty"`
	PhotoMimeType     *string     `json:"photo_mime_type,omitempty"`
	Timestamp         time.Time   `json:"timestamp"`            
	Status            string      `json:"status"`               
	IsReply           bool        `json:"is_reply"`             
	OriginalMessageID int64       `json:"original_message_id"`  
	IsForwarded       bool        `json:"is_forwarded"`       
	IsDeleted         bool        `json:"is_deleted"`
	Reactions         []Reaction  `json:"reactions"`
}

type MessageStatus struct {
	ID          int64  `json:"id"`
	MessageID   int64  `json:"message_id"`
	RecipientID int64  `json:"recipient_id"`
	IsReceived  bool   `json:"is_received"`
	IsRead      bool   `json:"is_read"`
}

type Reaction struct {
	MessageID int64  `json:"message_id"` 
	UserID    int64  `json:"user_id"`
	Emoticon  string `json:"emoticon"` 
}