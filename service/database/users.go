package database

import (
	"fmt"
)

func (db *appdbimpl) GetUsers() ([]User, error) {
	query := "SELECT id, username, photo_url FROM users ORDER BY username ASC"

	rows, err := db.c.Query(query)
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

		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("error during row iteration: %w", err)
		}
	}

	return users, nil
}
