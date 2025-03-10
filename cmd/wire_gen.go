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
	logrusLogger := ProvideLogrusLogger()
	observabilityLogger := ProvideLogrusLoggerImpl(logrusLogger)
	config, err := ProvideConfig()
	if err != nil {
		return nil, nil, err
	}
	client := ProvideMCPClient(observabilityLogger, config)
	toolsProvider, err := ProvideToolsProvider(client)
	if err != nil {
		return nil, nil, err
	}
	chatHistoryStorage := ProvideChatHistoryStorage()
	tracingService := ProvideTracingService(config, observabilityLogger)
	baseServer, err := ProvideMCPBaseServer(observabilityLogger)
	if err != nil {
		return nil, nil, err
	}
	authService, err := provideAuthenticator(ctx, config, observabilityLogger)
	if err != nil {
		return nil, nil, err
	}
	googleOAuthTokenSourceStorage := ProvideGoogleOAuthTokenSourceStorage(config)
	googleService := ProvideGoogleService(config, googleOAuthTokenSourceStorage)
	container := NewContainer(logger, client, toolsProvider, chatHistoryStorage, config, tracingService, logrusLogger, observabilityLogger, baseServer, authService, googleService, googleOAuthTokenSourceStorage)
	return container, func() {
	}, nil
}
