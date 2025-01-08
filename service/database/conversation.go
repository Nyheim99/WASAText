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