package database

import (
	"database/sql"
	"fmt"
)

func (db *appdbimpl) CreatePrivateConversation(userID, recipientID int64) (int64, error) {
	var existingConversationID int64

	err := db.c.QueryRow(`
		SELECT c.id 
		FROM conversations c
		JOIN conversation_participants cp1 ON c.id = cp1.conversation_id
		JOIN conversation_participants cp2 ON c.id = cp2.conversation_id
		WHERE c.conversation_type = 'private' 
		AND cp1.user_id = ? 
		AND cp2.user_id = ?
	`, userID, recipientID).Scan(&existingConversationID)

	if err == nil {
		return existingConversationID, fmt.Errorf("a private conversation between these users already exists")
	} else if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check for existing conversation: %w", err)
	}

	result, err := db.c.Exec(`
		INSERT INTO conversations (conversation_type) VALUES ('private')
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to create conversation: %w", err)
	}

	var conversationID int64

	conversationID, err = result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve conversation ID: %w", err)
	}

	_, err = db.c.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES (?, ?), (?, ?)
	`, conversationID, recipientID, conversationID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to add participants: %w", err)
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

func (db *appdbimpl) SetGroupName(conversationID int64, name string) error {
	_, err := db.c.Exec(
		`UPDATE conversations SET name = ? WHERE id = ? AND conversation_type = 'group'`,
		name, conversationID,
	)
	if err != nil {
		return fmt.Errorf("failed to update group name: %w", err)
	}
	return nil
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

// ConversationPreview represents a preview of a conversation for listing.
type ConversationPreview struct {
	ConversationID       int64   `json:"conversation_id"`
	ConversationType     string  `json:"conversation_type"`
	DisplayName          string  `json:"display_name"`
	DisplayPhotoURL      string  `json:"display_photo_url"`
	LastMessageID        int64   `json:"last_message_id"`
	LastMessageContent   *string `json:"last_message_content,omitempty"`
	LastMessageHasPhoto  bool    `json:"last_message_has_photo"`
	LastMessageTimestamp string  `json:"last_message_timestamp"`
	LastMessageSenderID  int64   `json:"last_message_sender_id,omitempty"`
	LastMessageSender    string  `json:"last_message_sender,omitempty"`
	LastMessageIsDeleted bool    `json:"last_message_is_deleted"`
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
			m.id AS last_message_id,  -- Retrieve last message ID
			m.content AS last_message_content,
			CASE WHEN m.photo_data IS NOT NULL THEN 1 ELSE 0 END AS last_message_has_photo,
			COALESCE(m.timestamp, '1970-01-01T00:00:00Z') AS last_message_timestamp,
			m.sender_id AS last_message_sender_id,
			sender.username AS last_message_sender,
			CASE WHEN m.is_deleted = 1 THEN 1 ELSE 0 END AS last_message_is_deleted
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
		LEFT JOIN 
			users sender ON sender.id = m.sender_id
		WHERE 
			cp.user_id = ?
		ORDER BY 
    	m.timestamp DESC
		LIMIT 50;
	`

	rows, err := db.c.Query(query, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []ConversationPreview

	for rows.Next() {
		var conversation ConversationPreview
		var lastMessageHasPhoto int
		var lastMessageSenderID sql.NullInt64
		var lastMessageSender sql.NullString
		var lastMessageIsDeleted int
		var lastMessageID int64

		if err := rows.Scan(
			&conversation.ConversationID,
			&conversation.ConversationType,
			&conversation.DisplayName,
			&conversation.DisplayPhotoURL,
			&lastMessageID,
			&conversation.LastMessageContent,
			&lastMessageHasPhoto,
			&conversation.LastMessageTimestamp,
			&lastMessageSenderID,
			&lastMessageSender,
			&lastMessageIsDeleted,
		); err != nil {
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}

		conversation.LastMessageHasPhoto = lastMessageHasPhoto == 1
		conversation.LastMessageIsDeleted = lastMessageIsDeleted == 1
		conversation.LastMessageID = lastMessageID

		if lastMessageSenderID.Valid {
			conversation.LastMessageSenderID = lastMessageSenderID.Int64
		}
		if lastMessageSender.Valid {
			conversation.LastMessageSender = lastMessageSender.String
		} else {
			conversation.LastMessageSender = "Unknown"
		}

		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

type ConversationDetails struct {
	ConversationID   int64     `json:"conversation_id"`
	ConversationType string    `json:"conversation_type"`
	DisplayName      string    `json:"display_name"`
	PhotoURL         string    `json:"display_photo_url"`
	Participants     []User    `json:"participants,omitempty"`
	Messages         []Message `json:"messages,omitempty"`
}

func (db *appdbimpl) GetConversation(conversationID int64) (*ConversationDetails, error) {
	var conversation ConversationDetails

	err := db.c.QueryRow(`
		SELECT id, conversation_type, name, photo_url
		FROM conversations
		WHERE id = ?`, conversationID).Scan(
		&conversation.ConversationID,
		&conversation.ConversationType,
		&conversation.DisplayName,
		&conversation.PhotoURL,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("conversation not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to retrieve conversation: %w", err)
	}

	// Fetch messages
	messageRows, err := db.c.Query(`
		SELECT 
			m.id, m.conversation_id, m.sender_id, u.username, 
			m.content, m.photo_data, m.photo_mime_type, m.timestamp, m.status, 
			m.is_reply, m.original_message_id, 
			m.is_forwarded, m.is_deleted
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.conversation_id = ?
		ORDER BY m.timestamp ASC`, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}
	defer messageRows.Close()

	messages := []Message{}
	for messageRows.Next() {
		var msg Message
		var photoData []byte
		var photoMimeType sql.NullString

		err := messageRows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.Content,
			&photoData,
			&photoMimeType,
			&msg.Timestamp,
			&msg.Status,
			&msg.IsReply,
			&msg.OriginalMessageID,
			&msg.IsForwarded,
			&msg.IsDeleted,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}

		if len(photoData) > 0 {
			msg.PhotoData = &photoData
		}
		if photoMimeType.Valid {
			msg.PhotoMimeType = &photoMimeType.String
		}

		// Fetch reactions for the message
		reactionRows, err := db.c.Query(`
			SELECT user_id, emoticon
			FROM reactions
			WHERE message_id = ?`, msg.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve reactions: %w", err)
		}
		defer reactionRows.Close()

		reactions := []Reaction{}
		for reactionRows.Next() {
			var reaction Reaction
			if err := reactionRows.Scan(&reaction.UserID, &reaction.Emoticon); err != nil {
				return nil, fmt.Errorf("failed to scan reaction: %w", err)
			}
			reactions = append(reactions, reaction)
		}
		msg.Reactions = reactions

		messages = append(messages, msg)
	}

	conversation.Messages = messages
	return &conversation, nil
}

func (db *appdbimpl) AddToGroup(conversationID int64, newParticipants []int64) error {
	if len(newParticipants) == 0 {
		return fmt.Errorf("no participants to add")
	}

	var conversationType string
	err := db.c.QueryRow(`
		SELECT conversation_type FROM conversations WHERE id = ?
	`, conversationID).Scan(&conversationType)

	if err == sql.ErrNoRows {
		return fmt.Errorf("conversation does not exist")
	} else if err != nil {
		return fmt.Errorf("failed to retrieve conversation: %w", err)
	}

	if conversationType != "group" {
		return fmt.Errorf("cannot add members to a private conversation")
	}

	query := "INSERT INTO conversation_participants (conversation_id, user_id) VALUES "
	args := []interface{}{}
	for _, userID := range newParticipants {
		query += "(?, ?),"
		args = append(args, conversationID, userID)
	}

	query = query[:len(query)-1]

	_, err = db.c.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to add participants: %w", err)
	}

	return nil
}

func (db *appdbimpl) LeaveGroup(conversationID int64, userID int64) error {
	// Step 1: Check if the conversation exists and is a group
	var conversationType string
	err := db.c.QueryRow(`
		SELECT conversation_type FROM conversations WHERE id = ?
	`, conversationID).Scan(&conversationType)

	if err == sql.ErrNoRows {
		return fmt.Errorf("conversation does not exist")
	} else if err != nil {
		return fmt.Errorf("failed to retrieve conversation: %w", err)
	}

	if conversationType != "group" {
		return fmt.Errorf("cannot leave a private conversation")
	}

	// Step 2: Remove the user from the conversation participants
	_, err = db.c.Exec(`
		DELETE FROM conversation_participants 
		WHERE conversation_id = ? AND user_id = ?
	`, conversationID, userID)

	if err != nil {
		return fmt.Errorf("failed to remove user from group: %w", err)
	}

	// Step 3: Check how many participants are left
	var participantCount int
	err = db.c.QueryRow(`
		SELECT COUNT(*) FROM conversation_participants WHERE conversation_id = ?
	`, conversationID).Scan(&participantCount)

	if err != nil {
		return fmt.Errorf("failed to check remaining participants: %w", err)
	}

	// Step 4: If only 1 participant remains, delete the group conversation
	if participantCount == 1 {
		_, err = db.c.Exec(`
			DELETE FROM conversations WHERE id = ?
		`, conversationID)

		if err != nil {
			return fmt.Errorf("failed to delete group conversation: %w", err)
		}

		// Also delete the last participant to completely remove the group
		_, err = db.c.Exec(`
			DELETE FROM conversation_participants WHERE conversation_id = ?
		`, conversationID)

		if err != nil {
			return fmt.Errorf("failed to remove last participant: %w", err)
		}
	}

	return nil
}
