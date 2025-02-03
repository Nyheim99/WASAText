package database

import (
	"database/sql"
	"fmt"
)

// SendMessage adds a message to a conversation, supporting both text and image messages.
func (db *appdbimpl) SendMessage(conversationID, senderID int64, content *string, photoData *[]byte, photoMimeType *string) (int64, error) {
	// Ensure that either content or photoData is provided, but not both
	if (content != nil && photoData != nil) || (content == nil && photoData == nil) {
		return 0, fmt.Errorf("a message must contain either text or an image, but not both")
	}

	// Prepare the query dynamically based on whether it's a text or photo message
	var result sql.Result
	var err error

	if content != nil {
		// Insert a text message
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, content, timestamp, status)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'sent')
		`, conversationID, senderID, *content)
	} else {
		// Insert an image message
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, photo_data, photo_mime_type, timestamp, status)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 'sent')
		`, conversationID, senderID, *photoData, *photoMimeType)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to add message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve message ID: %w", err)
	}

	// Update the last message ID in the conversations table
	_, err = db.c.Exec(`
		UPDATE conversations
		SET last_message_id = ?
		WHERE id = ?
	`, messageID, conversationID)
	if err != nil {
		return 0, fmt.Errorf("failed to update last message ID: %w", err)
	}

	return messageID, nil
}
