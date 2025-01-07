package database

import (
    "database/sql"
)

// GetUserByUsername retrieves a user by their username
func (db *appdbimpl) GetUserByUsername(username string) (int64, error) {
    var identifier int64
    err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&identifier)
    if err == sql.ErrNoRows {
        return 0, nil
    }
    return identifier, err
}

// CreateUser inserts a new user into the database
func (db *appdbimpl) CreateUser(username string) (int64, error) {
    result, err := db.c.Exec("INSERT INTO users (username) VALUES (?)", username)
    if err != nil {
        return 0, err
    }
    identifier, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }
    return identifier, nil
}

func (db *appdbimpl) DoesUserExist(userID int64) (bool, error) {
    var exists bool
    err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)", userID).Scan(&exists)
    return exists, err
}

func (db *appdbimpl) GetUser(userId int64) (*User, error) {
	var user User
	query := `SELECT id, username, photoUrl FROM users WHERE id = ?`
	err := db.c.QueryRow(query, userId).Scan(&user.UserID, &user.Username, &user.PhotoUrl)
	if err == sql.ErrNoRows {
		return nil, nil // User not found
	} else if err != nil {
		return nil, err
	}
	return &user, nil
}

func (db *appdbimpl) GetUserConversations(userID int64) ([]Conversation, error) {
	rows, err := db.c.Query(`
		SELECT c.id, c.name, c.conversationType, c.photoUrl, c.lastMessageId
		FROM conversations c
		JOIN conversation_participants cp ON c.id = cp.conversationId
		WHERE cp.userId = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ConversationID, &conv.Name, &conv.ConversationType, &conv.PhotoUrl, &conv.LastMessageID); err != nil {
			return nil, err
		}
		conversations = append(conversations, conv)
	}

    if conversations == nil {
    	conversations = []Conversation{}
	}	

	return conversations, nil
}

func (db *appdbimpl) DoesUsernameExist(username string) (bool, error) {
	var count int
	err := db.c.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, username).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (db *appdbimpl) SetMyUserName(userID int64, username string) error {
	_, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
	return err
}

func (db *appdbimpl) SetMyPhoto(userID int64, photoURL string) error {
	_, err := db.c.Exec(`UPDATE users SET photoUrl = ? WHERE id = ?`, photoURL, userID)
	return err
}