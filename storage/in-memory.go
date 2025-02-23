package storage

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
	"time"
)

// Implementation of the interface
type InMemoryChatHistoryStorage struct {
	conversations map[uuid.UUID]*ChatHistory
	mu            sync.RWMutex
}

func NewInMemoryChatHistoryStorage() *InMemoryChatHistoryStorage {
	return &InMemoryChatHistoryStorage{
		conversations: make(map[uuid.UUID]*ChatHistory),
	}
}

func (s *InMemoryChatHistoryStorage) CreateChat() (*ChatHistory, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat := &ChatHistory{
		UUID:      uuid.New(),
		Messages:  []Message{},
		CreatedAt: time.Now(),
	}

	s.conversations[chat.UUID] = chat
	return chat, nil
}

func (s *InMemoryChatHistoryStorage) AddMessage(chatID uuid.UUID, message Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	chat, exists := s.conversations[chatID]
	if !exists {
		return fmt.Errorf("chat with ID %s not found", chatID)
	}

	chat.Messages = append(chat.Messages, message)
	return nil
}

func (s *InMemoryChatHistoryStorage) GetChat(id uuid.UUID) (*ChatHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chat, exists := s.conversations[id]
	if !exists {
		return nil, fmt.Errorf("chat with ID %s not found", id)
	}

	return chat, nil
}

func (s *InMemoryChatHistoryStorage) ListChatHistories() ([]ChatHistory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	chats := make([]ChatHistory, 0, len(s.conversations))
	for _, chat := range s.conversations {
		chats = append(chats, *chat)
	}

	return chats, nil
}

func (s *InMemoryChatHistoryStorage) DeleteChat(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.conversations[id]; !exists {
		return fmt.Errorf("chat with ID %s not found", id)
	}

	delete(s.conversations, id)
	return nil
}
