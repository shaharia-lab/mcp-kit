// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package cmd

import (
	"context"
)

// Injectors from wire.go:

// InitializeAPI sets up the dependency injection
func InitializeAPI(ctx context.Context) (*Container, func(), error) {
	logger := ProvideLogger()
	config, err := ProvideConfig()
	if err != nil {
		return nil, nil, err
	}
	client := ProvideMCPClient(logger, config)
	toolsProvider, err := ProvideToolsProvider(client)
	if err != nil {
		return nil, nil, err
	}
	chatHistoryStorage := ProvideChatHistoryStorage()
	mux := ProvideRouter(ctx, client, logger, chatHistoryStorage, toolsProvider)
	container := NewContainer(logger, client, toolsProvider, chatHistoryStorage, mux, config)
	return container, func() {
	}, nil
}
