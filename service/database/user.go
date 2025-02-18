package database

import (
	"database/sql"
	"errors"
	"fmt"
)

//Get a user's identifer with their username
func (db *appdbimpl) GetUserByUsername(username string) (int64, error) {
	var identifier int64
	err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&identifier)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return identifier, err
}

//Create a user with their username and return the identifier
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

//Check if the username is already taken
func (db *appdbimpl) DoesUsernameExist(username string) (bool, error) {
	var exists bool
	err := db.c.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&exists)
	return exists, err
}

//Get a full user object using their id
func (db *appdbimpl) GetUser(userId int64) (*User, error) {
	var user User
	query := `SELECT id, username, photo_url FROM users WHERE id = ?`
	err := db.c.QueryRow(query, userId).Scan(&user.ID, &user.Username, &user.PhotoURL)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}
	return &user, nil
}

//Updates a user's username
func (db *appdbimpl) SetMyUserName(userID int64, username string) error {
	_, err := db.c.Exec(`UPDATE users SET username = ? WHERE id = ?`, username, userID)
	if err != nil {
		return fmt.Errorf("error updating username: %w", err)
	}
	return nil
}

//Updates a user's profile picture
func (db *appdbimpl) SetMyPhoto(userID int64, photoURL string) error {
	_, err := db.c.Exec(`UPDATE users SET photo_url = ? WHERE id = ?`, photoURL, userID)
	if err != nil {
		return fmt.Errorf("error updating profile photo: %w", err)
	}
	return nil
}
