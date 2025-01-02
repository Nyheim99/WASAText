package database

// GetUserByUsername retrieves a user by their username
func (db *appdbimpl) GetUserByUsername(username string) (int64, error) {
    var identifier int64
    err := db.c.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&identifier)
    return identifier, err
}

// CreateUser inserts a new user into the database
func (db *appdbimpl) CreateUser(username string) (int64, error) {
    result, err := db.c.Exec("INSERT INTO users (username) VALUES (?)", username)
    identifier, err := result.LastInsertId()
    return identifier, err
}

