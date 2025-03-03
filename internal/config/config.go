package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/shaharia-lab/mcp-kit/internal/auth"
	"time"
)

type Config struct {
	APIServerPort int      `envconfig:"API_SERVER_PORT" default:"8081"`
	MCPServerURL  string   `envconfig:"MCP_SERVER_URL" default:"http://localhost:8080/events"`
	MCPServerPort int      `envconfig:"MCP_SERVER_PORT" default:"8080"`
	ToolsEnabled  []string `envconfig:"TOOLS_ENABLED" default:"get_weather"`
	Tracing       TracingConfig
	Auth          auth.Config
}

// TracingConfig holds the configuration for the tracing service
type TracingConfig struct {
	Enabled         bool          `envconfig:"TRACING_ENABLED" default:"false"`
	ServiceName     string        `envconfig:"TRACING_SERVICE_NAME" default:"service"`
	EndpointAddress string        `envconfig:"TRACING_ENDPOINT_ADDRESS" default:"localhost:4317"`
	Timeout         time.Duration `envconfig:"TRACING_TIMEOUT" default:"5s"`
	SamplingRate    float64       `envconfig:"TRACING_SAMPLING_RATE" default:"1.0"`
	BatchTimeout    time.Duration `envconfig:"TRACING_BATCH_TIMEOUT" default:"5s"`
	Environment     string        `envconfig:"TRACING_ENVIRONMENT" default:"development"`
	Version         string        `envconfig:"TRACING_VERSION" default:"0.1.0"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
