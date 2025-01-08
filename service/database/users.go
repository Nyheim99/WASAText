package database

import (
    "database/sql"
    "fmt"
)

func (db *appdbimpl) GetUsers(conversationID *int64) ([]User, error) {
    var rows *sql.Rows
    var err error

    if conversationID != nil {
        rows, err = db.c.Query(`
            SELECT u.id, u.username, u.photo_url 
            FROM users u
            JOIN conversation_participants cp ON u.id = cp.user_id
            WHERE cp.conversation_id = ?`, *conversationID)
    } else {
        rows, err = db.c.Query(`SELECT id, username, photo_url FROM users`)
    }

    if err != nil {
        return nil, fmt.Errorf("error retrieving users: %w", err)
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Username, &user.PhotoURL); err != nil {
            return nil, fmt.Errorf("error scanning user: %w", err)
        }
        users = append(users, user)
    }

    if users == nil {
        return []User{}, nil
    }

    return users, nil
}
