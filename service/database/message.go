package database

import (
	"database/sql"
	"fmt"
	"errors"
)

func (db *appdbimpl) SendMessage(conversationID, senderID int64, content *string, photoData *[]byte, photoMimeType *string, originalMessageID int64) (int64, error) {
	if (content != nil && photoData != nil) || (content == nil && photoData == nil) {
		return 0, fmt.Errorf("a message must contain either text or an image, but not both")
	}

	isReply := originalMessageID > 0
	var result sql.Result
	var err error

	if content != nil {
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, content, timestamp, status, is_reply, original_message_id)
			VALUES (?, ?, ?, CURRENT_TIMESTAMP, 'sent', ?, ?)
		`, conversationID, senderID, *content, isReply, originalMessageID)
	} else {
		result, err = db.c.Exec(`
			INSERT INTO messages (conversation_id, sender_id, photo_data, photo_mime_type, timestamp, status, is_reply, original_message_id)
			VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, 'sent', ?, ?)
		`, conversationID, senderID, *photoData, *photoMimeType, isReply, originalMessageID)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to add message: %w", err)
	}

	messageID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve message ID: %w", err)
	}

	_, err = db.c.Exec(`
		INSERT INTO message_status (message_id, user_id, is_read)
		SELECT ?, user_id, CASE WHEN user_id = ? THEN TRUE ELSE FALSE END
		FROM conversation_participants
		WHERE conversation_id = ?
	`, messageID, senderID, conversationID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert message status for participants: %w", err)
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

	_, err = db.c.Exec(`
	DELETE FROM message_status WHERE message_id = ?
`, messageID)
	if err != nil {
		return fmt.Errorf("failed to delete message status: %w", err)
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
		if errors.Is(err, sql.ErrNoRows) {
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
		INSERT INTO message_status (message_id, user_id, is_read)
		SELECT ?, user_id, CASE WHEN user_id = ? THEN TRUE ELSE FALSE END
		FROM conversation_participants
		WHERE conversation_id = ?
	`, messageID, senderID, conversationID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert message status for participants: %w", err)
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

func (db *appdbimpl) MarkMessagesAsRead(conversationID, userID int64) error {
	_, err := db.c.Exec(`
		UPDATE message_status 
		SET is_read = TRUE 
		WHERE message_id IN (
			SELECT id FROM messages WHERE conversation_id = ?
		) AND user_id = ?
	`, conversationID, userID)
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	// Optionally, check if all participants have read all messages and update the message status to 'read'
	_, err = db.c.Exec(`
		UPDATE messages 
		SET status = 'read'
		WHERE id IN (
			SELECT message_id 
			FROM message_status 
			WHERE conversation_id = ? 
			GROUP BY message_id 
			HAVING COUNT(user_id) = (SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ?)
		)
	`, conversationID, conversationID)
	if err != nil {
		return fmt.Errorf("failed to update message status to read: %w", err)
	}

	return nil
}
