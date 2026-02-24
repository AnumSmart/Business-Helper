package domain

import "time"

// Внутренние структуры для БД (не protobuf)
type Message struct {
	ID        int64
	MessageID int64
	ChatID    int64
	UserID    int64
	Text      string
	Direction string // "incoming" или "outgoing"
	Status    string // "sent", "pending", "failed"
	Timestamp time.Time
}

type CallbackLog struct {
	ID         int64
	CallbackID string
	UserID     int64
	ChatID     int64
	MessageID  int64
	Data       string
	Timestamp  time.Time
}
