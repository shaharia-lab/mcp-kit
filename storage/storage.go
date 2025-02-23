package storage

import (
	"github.com/shaharia-lab/goai"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	goai.LLMMessage
	GeneratedAt time.Time `json:"generated_at"`
}

type ChatHistory struct {
	UUID      uuid.UUID `json:"uuid"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatHistoryStorage defines the interface for conversation history storage
type ChatHistoryStorage interface {
	// CreateChat initializes a new chat conversation
	CreateChat() (*ChatHistory, error)

	// AddMessage adds a new message to an existing conversation
	AddMessage(chatID uuid.UUID, message Message) error

	// GetChat retrieves a conversation by its UUID
	GetChat(id uuid.UUID) (*ChatHistory, error)

	// ListChatHistories returns all stored conversations
	ListChatHistories() ([]ChatHistory, error)

	// DeleteChat removes a conversation by its UUID
	DeleteChat(id uuid.UUID) error
}
