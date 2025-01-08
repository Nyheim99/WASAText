package database

import (
    "database/sql"
		"fmt"
)

func (db *appdbimpl) GetUserByUsername(username string) (int64, error) {
    var identifier int64
    err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&identifier)
    if err == sql.ErrNoRows {
        return 0, nil
    }
    return identifier, err
}

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

func (db *appdbimpl) DoesUsernameExist(username string) (bool, error) {
    var exists bool
    err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
    return exists, err
}

func (db *appdbimpl) GetUser(userId int64) (*User, error) {
    var user User
    query := `SELECT id, username, photo_url FROM users WHERE id = ?`
    err := db.c.QueryRow(query, userId).Scan(&user.ID, &user.Username, &user.PhotoURL)
    if err == sql.ErrNoRows {
        return nil, nil
    } else if err != nil {
        return nil, fmt.Errorf("error retrieving user: %w", err)
    }
    return &user, nil
}

func (db *appdbimpl) GetUserConversations(userID int64) ([]Conversation, error) {
    rows, err := db.c.Query(`
        SELECT c.id, c.name, c.conversation_type, c.photo_url, c.last_message_id
        FROM conversations c
        JOIN conversation_participants cp ON c.id = cp.conversation_id
        WHERE cp.user_id = ?`, userID)
    if err != nil {
        return nil, fmt.Errorf("error retrieving user conversations: %w", err)
    }
    defer rows.Close()

    var conversations []Conversation
    for rows.Next() {
        var conv Conversation
        if err := rows.Scan(&conv.ID, &conv.Name, &conv.ConversationType, &conv.PhotoURL, &conv.LastMessageID); err != nil {
            return nil, fmt.Errorf("error scanning conversation: %w", err)
        }
        conversations = append(conversations, conv)
    }

    return conversations, nil
}

func (db *appdbimpl) SetMyUserName(userID int64, username string) error {
    _, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
    if err != nil {
        return fmt.Errorf("error updating username: %w", err)
    }
    return nil
}

func (db *appdbimpl) SetMyPhoto(userID int64, photoURL string) error {
    _, err := db.c.Exec(`UPDATE users SET photo_url = ? WHERE id = ?`, photoURL, userID)
    if err != nil {
        return fmt.Errorf("error updating profile photo: %w", err)
    }
    return nil
}