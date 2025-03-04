//go:build wireinject
// +build wireinject

package cmd

import (
	"context"
	"github.com/google/wire"
)

// InitializeAPI sets up the dependency injection
func InitializeAPI(ctx context.Context) (*Container, func(), error) {
	panic(wire.Build(
		NewContainer,
		ProvideLogger,
		ProvideMCPClient,
		ProvideToolsProvider,
		ProvideChatHistoryStorage,
		ProvideConfig,
		ProvideTracingService,
		ProvideLogrusLogger,
		ProvideLogrusLoggerImpl,
		ProvideMCPBaseServer,
		provideAuthenticator,
	))
}
