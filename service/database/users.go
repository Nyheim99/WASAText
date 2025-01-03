package database

import "database/sql"

func (db *appdbimpl) GetUsers(conversationID *int64) ([]User, error) {
    var rows *sql.Rows
    var err error

    if conversationID != nil {
        rows, err = db.c.Query(`
            SELECT u.id, u.username, u.photoUrl 
            FROM users u
            JOIN conversation_participants cp ON u.id = cp.userId
            WHERE cp.conversationId = ?`, *conversationID)
    } else {
        rows, err = db.c.Query(`SELECT id, username, photoUrl FROM users`)
    }

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.UserID, &user.Username, &user.PhotoUrl); err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    return users, nil
}