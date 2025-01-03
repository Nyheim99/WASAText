package database

import "database/sql"

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
