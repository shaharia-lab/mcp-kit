package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shaharia-lab/mcp-kit/storage"
	"log"
	"net/http"
)

// Handler to list all chats
func ChatHistoryListsHandler(logger *log.Logger, historyStorage storage.ChatHistoryStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fetch all chat histories from storage
		chats, err := historyStorage.ListChatHistories()
		if err != nil {
			logger.Printf("Failed to retrieve chat histories: %v", err)
			http.Error(w, `{"error": "Failed to retrieve chat histories"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		response := struct {
			Chats []storage.ChatHistory `json:"chats"`
		}{
			Chats: chats,
		}

		// Encode response and handle potential errors
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Printf("Error encoding chat list response: %v", err)
			http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}

// Handler to get a single chat by chatId
func GetChatHandler(logger *log.Logger, historyStorage storage.ChatHistoryStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract 'chatId' from URL parameters
		chatUUID := chi.URLParam(r, "chatId")
		if chatUUID == "" {
			http.Error(w, `{"error": "Chat ID is required"}`, http.StatusBadRequest)
			return
		}

		// Parse the provided Chat ID as UUID
		parsedChatUUID, err := uuid.Parse(chatUUID)
		if err != nil {
			logger.Printf("Invalid chat UUID provided: %s", chatUUID)
			http.Error(w, `{"error": "Invalid chat ID"}`, http.StatusBadRequest)
			return
		}

		// Fetch the chat from storage by its UUID
		chat, err := historyStorage.GetChat(parsedChatUUID)
		if err != nil {
			logger.Printf("Chat not found for UUID: %v, error: %v", parsedChatUUID, err)
			http.Error(w, `{"error": "Chat not found"}`, http.StatusNotFound)
			return
		}

		// Encode and return the chat as a JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(chat); err != nil {
			logger.Printf("Error encoding chat response: %v", err)
			http.Error(w, `{"error": "Failed to encode response"}`, http.StatusInternalServerError)
		}
	}
}
