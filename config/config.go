package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	MCPServerURL  string   `envconfig:"MCP_SERVER_URL" default:"http://localhost:8080/events"`
	MCPServerPort int      `envconfig:"MCP_SERVER_PORT" default:"8080"`
	ToolsEnabled  []string `envconfig:"TOOLS_ENABLED" default:"get_weather"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
