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
		var conversation ConversationPreview
		if err := rows.Scan(
			&conversation.ConversationID,
			&conversation.ConversationType,
			&conversation.DisplayName,
			&conversation.DisplayPhotoURL,
			&conversation.LastMessageContent,
			&conversation.LastMessagePhotoURL,
			&conversation.LastMessageTimestamp,
		); err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			return nil, fmt.Errorf("failed to scan conversation row: %w", err)
		}
		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		fmt.Printf("Row iteration error: %v\n", err)
		return nil, fmt.Errorf("error iterating through rows: %w", err)
	}

	return conversations, nil
}

type ConversationDetails struct {
	ConversationID   int64  `json:"conversation_id"`
	ConversationType string `json:"conversation_type"`
	DisplayName      string `json:"display_name"`
	PhotoURL         string `json:"display_photo_url"`
	Participants     []User `json:"participants,omitempty"`
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
		return nil, fmt.Errorf("failed to retrieve conversationnn: %w", err)
	}

	if conversation.ConversationType == "group" {
		rows, err := db.c.Query(`
			SELECT users.id, users.username, users.photo_url
			FROM conversation_participants
			JOIN users ON conversation_participants.user_id = users.id
			WHERE conversation_participants.conversation_id = ?`, conversationID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve participants: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var participant User
			if err := rows.Scan(&participant.ID, &participant.Username, &participant.PhotoURL); err != nil {
				return nil, fmt.Errorf("failed to scan participant: %w", err)
			}
			conversation.Participants = append(conversation.Participants, participant)
		}
	}

	// Step 3: Fetch all messages in the conversation, including sender username
	messageRows, err := db.c.Query(`
		SELECT 
			m.id, m.conversation_id, m.sender_id, u.username, 
			m.content, m.photo_url, m.timestamp, m.status, 
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

	// Step 4: Populate the messages array
	messages := []Message{}
	for messageRows.Next() {
		var msg Message
		err := messageRows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.SenderUsername,
			&msg.Content,
			&msg.PhotoURL,
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

		// Step 5: Fetch reactions for this message
		reactionRows, err := db.c.Query(`
			SELECT r.user_id, u.username, r.emoticon
			FROM reactions r
			JOIN users u ON r.user_id = u.id
			WHERE r.message_id = ?`, msg.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve reactions: %w", err)
		}

		for reactionRows.Next() {
			var reaction Reaction
			if err := reactionRows.Scan(&reaction.UserID, &reaction.Username, &reaction.Emoticon); err != nil {
				return nil, fmt.Errorf("failed to scan reaction: %w", err)
			}
			msg.Reactions = append(msg.Reactions, reaction)
		}
		reactionRows.Close()

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
 
