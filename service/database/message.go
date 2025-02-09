package database

import (
	"database/sql"
	"fmt"
)

func (db *appdbimpl) SendMessage(conversationID, senderID int64, content *string, photoData *[]byte, photoMimeType *string) (int64, error) {
	if (content != nil && photoData != nil) || (content == nil && photoData == nil) {
		return 0, fmt.Errorf("a message must contain either text or an image, but not both")
	}

	var result sql.Result
	var err error

	if content != nil {
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, content, timestamp, status)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'sent')
		`, conversationID, senderID, *content)
	} else {
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

func (db *appdbimpl) DeleteMessage(conversationID, messageID, userID int64) error {
	var count int
	err := db.c.QueryRow(`
		SELECT COUNT(*) FROM messages 
		WHERE id = ? AND conversation_id = ? AND sender_id = ? AND is_deleted = FALSE
	`, messageID, conversationID, userID).Scan(&count)

	if err != nil {
		return fmt.Errorf("failed to check message existence: %w", err)
	}

	if count == 0 {
		return fmt.Errorf("message not found or already deleted")
	}

	_, err = db.c.Exec(`
		UPDATE messages 
		SET is_deleted = TRUE
		WHERE id = ? AND conversation_id = ? AND sender_id = ?
	`, messageID, conversationID, userID)

	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func (db *appdbimpl) CommentMessage(messageID, userID int64, emoticon string) error {
	_, err := db.c.Exec(`
		INSERT INTO reactions (message_id, user_id, emoticon)
		VALUES (?, ?, ?)
		ON CONFLICT (message_id, user_id) 
		DO UPDATE SET emoticon = excluded.emoticon
	`, messageID, userID, emoticon)

	if err != nil {
		return fmt.Errorf("failed to add reaction: %w", err)
	}

	return nil
}

func (db *appdbimpl) UncommentMessage(messageID, userID int64) error {
	_, err := db.c.Exec(`
		DELETE FROM reactions WHERE message_id = ? AND user_id = ?
	`, messageID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove reaction: %w", err)
	}

	return nil
}

func (db *appdbimpl) ForwardMessage(conversationID, senderID, originalMessageID int64) (int64, error) {
	var content sql.NullString
	var photoData []byte
	var photoMimeType sql.NullString
	err := db.c.QueryRow(`
		SELECT content, photo_data, photo_mime_type 
		FROM messages 
		WHERE id = ? AND is_deleted = FALSE
	`, originalMessageID).Scan(&content, &photoData, &photoMimeType)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("original message not found")
		}
		return 0, fmt.Errorf("failed to retrieve original message: %w", err)
	}

	var result sql.Result
	if content.Valid {
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, content, timestamp, is_forwarded, original_message_id)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, TRUE, ?)
		`, conversationID, senderID, content.String, originalMessageID)
	} else {
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, photo_data, photo_mime_type, timestamp, is_forwarded, original_message_id)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, TRUE, ?)
		`, conversationID, senderID, photoData, photoMimeType.String, originalMessageID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to forward message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve new message ID: %w", err)
	}

	_, err = db.c.Exec(`
		UPDATE conversations
		SET last_message_id = ?
		WHERE id = ?
	`, messageID, conversationID)
	if err != nil {
		return 0, fmt.Errorf("failed to update last message ID for conversation: %w", err)
	}

	return messageID, nil
}


