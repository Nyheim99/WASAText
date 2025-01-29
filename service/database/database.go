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
	DoesUsernameExist(username string) (bool, error)

	GetUser(userId int64) (*User, error)
	GetUsers() ([]User, error)
	
	SetMyUserName(userID int64, username string) error
	SetMyPhoto(userID int64, photoURL string) error
	
	GetOrCreatePrivateConversation(currentUserID, recipientID int64) (int64, error)
	CreateGroupConversation(creatorID int64, name, photoURL string, participants []int64) (int64, error)
	AddMessage(conversationID, senderID int64, content string) (int64, error)

	SetGroupPhoto(conversationID int64, photoURL string) error

	GetMyConversations(userID int64) ([]ConversationPreview, error)

	Ping() error
}

type appdbimpl struct {
	c *sql.DB
}

// New returns a new instance of AppDatabase based on the SQLite connection `db`.
// `db` is required - an error will be returned if `db` is `nil`.
func New(db *sql.DB) (AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}

	sqlStmts := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL CHECK (LENGTH(username) BETWEEN 3 AND 16),
			photo_url TEXT DEFAULT ''
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT DEFAULT '',
			conversation_type TEXT CHECK(conversation_type IN ('private', 'group')) NOT NULL,
			photo_url TEXT DEFAULT '',
			last_message_id INTEGER,
			FOREIGN KEY (last_message_id) REFERENCES messages(id)
		);`,
		`CREATE TABLE IF NOT EXISTS conversation_participants (
			conversation_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			PRIMARY KEY (conversation_id, user_id),
			FOREIGN KEY (conversation_id) REFERENCES conversations(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id INTEGER NOT NULL,
			sender_id INTEGER NOT NULL,
			content TEXT DEFAULT '',
			photo_url TEXT DEFAULT '',
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			status TEXT CHECK(status IN ('sent', 'received', 'read')) DEFAULT 'sent',
			is_reply BOOLEAN DEFAULT FALSE,
			original_message_id INTEGER,
			is_forwarded BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (conversation_id) REFERENCES conversations(id),
			FOREIGN KEY (sender_id) REFERENCES users(id),
			FOREIGN KEY (original_message_id) REFERENCES messages(id)
		);`,
		`CREATE TABLE IF NOT EXISTS reactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			emoticon TEXT NOT NULL, -- Emoji reactions
			FOREIGN KEY (message_id) REFERENCES messages(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
	}

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
