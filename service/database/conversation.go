package database

import (
	"database/sql"
	"fmt"
)

// GetOrCreatePrivateConversation retrieves an existing private conversation or creates a new one between two users.
func (db *appdbimpl) GetOrCreatePrivateConversation(currentUserID, recipientID int64) (int64, error) {
	var conversationID int64

	// Step 1: Try to fetch an existing private conversation
	err := db.c.QueryRow(`
		SELECT id
		FROM conversations
		WHERE conversation_type = 'private'
		AND id IN (
			SELECT conversation_id
			FROM conversation_participants
			WHERE user_id = ?
		)
		AND id IN (
			SELECT conversation_id
			FROM conversation_participants
			WHERE user_id = ?
		)
		LIMIT 1
	`, recipientID, currentUserID).Scan(&conversationID)

	if err == sql.ErrNoRows {
		// Step 2: No existing conversation; create a new one
		fmt.Println("No existing conversation found, creating a new one...")
		result, err := db.c.Exec(`
			INSERT INTO conversations (conversation_type)
			VALUES ('private')
		`)
		if err != nil {
			return 0, fmt.Errorf("failed to create conversation: %w", err)
		}

		conversationID, err = result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to retrieve conversation ID: %w", err)
		}

		// Step 3: Add participants to the new conversation
		_, err = db.c.Exec(`
			INSERT INTO conversation_participants (conversation_id, user_id)
			VALUES (?, ?), (?, ?)
		`, conversationID, recipientID, conversationID, currentUserID)
		if err != nil {
			return 0, fmt.Errorf("failed to add participants: %w", err)
		}
	} else if err != nil {
		return 0, fmt.Errorf("failed to query existing conversation: %w", err)
	}

	return conversationID, nil
}

// CreateGroupConversation creates a new group conversation with the given participants.
func (db *appdbimpl) CreateGroupConversation(creatorID int64, name, photoURL string, participants []int64) (int64, error) {
	// Step 1: Insert into the conversations table
	result, err := db.c.Exec(`
		INSERT INTO conversations (conversation_type, name, photo_url)
		VALUES ('group', ?, ?)
	`, name, photoURL)
	if err != nil {
		return 0, fmt.Errorf("failed to create group conversation: %w", err)
	}

	conversationID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve conversation ID: %w", err)
	}

	// Step 2: Add participants to the conversation, including the creator
	participantValues := ""
	args := []interface{}{}
	for _, participantID := range participants {
		participantValues += "(?, ?),"
		args = append(args, conversationID, participantID)
	}
	// Add the creator as a participant
	participantValues += "(?, ?)"
	args = append(args, conversationID, creatorID)

	_, err = db.c.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES `+participantValues, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to add participants: %w", err)
	}

	return conversationID, nil
}

// ConversationPreview represents a preview of a conversation for listing.
type ConversationPreview struct {
	ConversationID      int64  `json:"conversation_id"`
	ConversationType    string `json:"conversation_type"`
	DisplayName         string `json:"display_name"`
	DisplayPhotoURL     string `json:"display_photo_url"`
	LastMessageContent  string `json:"last_message_content,omitempty"`
	LastMessagePhotoURL string `json:"last_message_photo_url,omitempty"`
	LastMessageTimestamp string `json:"last_message_timestamp"`
}

func (db *appdbimpl) SetGroupPhoto(conversationID int64, photoURL string) error {
	_, err := db.c.Exec(
		`UPDATE conversations SET photo_url = $1 WHERE id = $2`,
		photoURL, conversationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group photo: %w", err)
	}
	return nil
}

// GetMyConversations retrieves a list of conversations for a given user, sorted by the latest message timestamp.
func (db *appdbimpl) GetMyConversations(userID int64) ([]ConversationPreview, error) {
	query := `
		SELECT 
			c.id AS conversation_id,
			c.conversation_type,
			CASE
				WHEN c.conversation_type = 'private' THEN u.username
				ELSE c.name
			END AS display_name,
			CASE
				WHEN c.conversation_type = 'private' THEN u.photo_url
				ELSE c.photo_url
			END AS display_photo_url,
			COALESCE(m.content, '') AS last_message_content,
			COALESCE(m.photo_url, '') AS last_message_photo_url,
			COALESCE(m.timestamp, '1970-01-01T00:00:00Z') AS last_message_timestamp
		FROM 
			conversations c
		JOIN 
			conversation_participants cp ON c.id = cp.conversation_id
		LEFT JOIN 
			messages m ON c.last_message_id = m.id
		LEFT JOIN 
			users u ON u.id = (
				SELECT cp2.user_id
				FROM conversation_participants cp2
				WHERE cp2.conversation_id = c.id AND cp2.user_id != ?
				LIMIT 1
			)
		WHERE 
			cp.user_id = ?
		ORDER BY 
			m.timestamp DESC;
	`

	rows, err := db.c.Query(query, userID, userID)
	if err != nil {
		fmt.Printf("Query error: %v\n", err)
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationPreview

	for rows.Next() {
		var conv ConversationPreview
		if err := rows.Scan(
			&conv.ConversationID,
			&conv.ConversationType,
			&conv.DisplayName,
			&conv.DisplayPhotoURL,
			&conv.LastMessageContent,
			&conv.LastMessagePhotoURL,
			&conv.LastMessageTimestamp,
		); err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Row iteration error: %v\n", err)
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return conversations, nil
}
