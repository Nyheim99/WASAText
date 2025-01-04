/*
Package database is the middleware between the app database and the code. All data (de)serialization (save/load) from a
persistent database are handled here. Database specific logic should never escape this package.

To use this package you need to apply migrations to the database if needed/wanted, connect to it (using the database
data source name from config), and then initialize an instance of AppDatabase from the DB connection.

For example, this code adds a parameter in `webapi` executable for the database data source name (add it to the
main.WebAPIConfiguration structure):

	DB struct {
		Filename string `conf:""`
	}

This is an example on how to migrate the DB and connect to it:

	// Start Database
	logger.Println("initializing database support")
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		logger.WithError(err).Error("error opening SQLite DB")
		return fmt.Errorf("opening SQLite: %w", err)
	}
	defer func() {
		logger.Debug("database stopping")
		_ = db.Close()
	}()

Then you can initialize the AppDatabase and pass it to the api package.
*/
package database

import (
	"database/sql"
	"errors"
	"fmt"
)

// AppDatabase is the high level interface for the DB
type AppDatabase interface {
	GetUserByUsername(username string) (int64, error)
	CreateUser(username string) (int64, error)
	DoesUserExist(userID int64) (bool, error)
	GetUser(userId int64) (*User, error)
	GetUsers(conversationID *int64) ([]User, error)
	GetUserConversations(userID int64) ([]Conversation, error)
	DoesUsernameExist(username string) (bool, error)
	UpdateUserName(userID int64, username string) error

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building a AppDatabase")
	}

	// Table creation SQL scripts
	sqlStmts := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL CHECK (LENGTH(username) BETWEEN 3 AND 16),
			photoUrl TEXT DEFAULT ''
		);`,
		// Messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			userId INTEGER NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT CHECK(status IN ('sent', 'read')) NOT NULL,
			isReply BOOLEAN DEFAULT FALSE,
			originalMessageId INTEGER,
			isForwarded BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (userId) REFERENCES users(id),
			FOREIGN KEY (originalMessageId) REFERENCES messages(id)
		);`,
		// Conversations table
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			conversationType TEXT CHECK(conversationType IN ('group', 'private')) NOT NULL,
			photoUrl TEXT,
			lastMessageId INTEGER,
			FOREIGN KEY (lastMessageId) REFERENCES messages(id)
		);`,
		// Conversation Participants table
		`CREATE TABLE IF NOT EXISTS conversation_participants (
			conversationId INTEGER NOT NULL,
			userId INTEGER NOT NULL,
			PRIMARY KEY (conversationId, userId),
			FOREIGN KEY (conversationId) REFERENCES conversations(id),
			FOREIGN KEY (userId) REFERENCES users(id)
		);`,
		// Comments table
		`CREATE TABLE IF NOT EXISTS comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			messageId INTEGER NOT NULL,
			userId INTEGER NOT NULL,
			emoticon TEXT,
			content TEXT,
			FOREIGN KEY (messageId) REFERENCES messages(id),
			FOREIGN KEY (userId) REFERENCES users(id)
		);`,
	}

	// Execute each SQL statement
	for _, sqlStmt := range sqlStmts {
		_, err := db.Exec(sqlStmt)
		if err != nil {
			return nil, fmt.Errorf("error creating database structure: %w", err)
		}
	}

	return &appdbimpl{c: db}, nil
}

func (db *appdbimpl) Ping() error {
	return db.c.Ping()
}
