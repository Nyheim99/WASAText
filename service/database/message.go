package database

import (
		"fmt"
)

// AddMessage adds a message to a conversation.
func (db *appdbimpl) AddMessage(conversationID, senderID int64, content string) (int64, error) {
	// Insert the message into the messages table
	result, err := db.c.Exec(`
		INSERT INTO messages (conversation_id, sender_id, content, timestamp, status)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'sent')
	`, conversationID, senderID, content)
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
