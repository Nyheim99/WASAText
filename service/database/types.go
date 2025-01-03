package database

type User struct {
	UserID        int64   `json:"userId"`
	Username      string  `json:"username"`
	PhotoUrl      string  `json:"photoUrl"`
}

type Comment struct {
	CommentID int64  `json:"commentId" db:"id"`
	MessageID int64  `json:"messageId" db:"messageId"`
	UserID   string `json:"userId"`
	Emoticon string `json:"emoticon"`
}

type Message struct {
	MessageID        int64     `json:"messageId"`
	UserID           int64     `json:"userId"`
	Content          string    `json:"content"`
	Timestamp        string    `json:"timestamp"`
	Status           string    `json:"status"`
	IsReply          bool      `json:"isReply"`
	OriginalMessageID int64    `json:"originalMessageId"`
	IsForwarded      bool      `json:"isForwarded"`
	Comments         []Comment `json:"comments"`
}

type Conversation struct {
	ConversationID   int64   `json:"conversationId"`
	Name             string  `json:"name"`
	ConversationType string  `json:"conversationType"`
	Users            []int64 `json:"users"`
	LastMessageID    int64   `json:"lastMessageId"`
	PhotoUrl         string  `json:"photoUrl"`
}
