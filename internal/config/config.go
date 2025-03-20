package config

import (
	"fmt"
	"github.com/shaharia-lab/mcp-kit/internal/tools"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	APIServerPort       int                `mapstructure:"api_server_port"`
	MCPServerURL        string             `mapstructure:"mcp_server_url"`
	MCPServerPort       int                `mapstructure:"mcp_server_port"`
	ToolsEnabled        []string           `mapstructure:"tools_enabled"`
	Tracing             TracingConfig      `mapstructure:"tracing"`
	Auth                AuthConfig         `mapstructure:"auth"`
	GoogleServiceConfig GoogleConfig       `mapstructure:"google"`
	Tools               *tools.ToolsConfig `mapstructure:"tools"`
}

// TracingConfig holds the configuration for the tracing service
type TracingConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	ServiceName     string        `mapstructure:"service_name"`
	EndpointAddress string        `mapstructure:"endpoint_address"`
	Timeout         time.Duration `mapstructure:"timeout"`
	SamplingRate    float64       `mapstructure:"sampling_rate"`
	BatchTimeout    time.Duration `mapstructure:"batch_timeout"`
	Environment     string        `mapstructure:"environment"`
	Version         string        `mapstructure:"version"`
}

// AuthConfig definition (moved from internal/auth)
type AuthConfig struct {
	AuthDomain       string        `mapstructure:"domain"`
	AuthClientID     string        `mapstructure:"client_id"`
	AuthClientSecret string        `mapstructure:"client_secret"`
	AuthCallbackURL  string        `mapstructure:"callback_url"`
	AuthTokenTTL     time.Duration `mapstructure:"token_ttl"`
	AuthAudience     string        `mapstructure:"audience"`
}

// GoogleConfig definition (moved from internal/service/google)
type GoogleConfig struct {
	ClientID        string   `mapstructure:"client_id"`
	ClientSecret    string   `mapstructure:"client_secret"`
	RedirectURL     string   `mapstructure:"redirect_url"`
	Scopes          []string `mapstructure:"scopes"`
	StateCookie     string   `mapstructure:"state_cookie"`
	TokenSourceFile string   `mapstructure:"token_source_file"`
	Enabled         bool     `mapstructure:"enabled"`
}

func Load(configFile string) (*Config, error) {
	var cfg Config

	setDefaults()
	tools.SetDefaults(viper.GetViper())

	// Configure Viper
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configFile)

	// Configure Viper to read environment variables
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read the config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal into struct
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Validate tools configuration
	if err := cfg.Tools.Validate(); err != nil {
		return nil, fmt.Errorf("tools validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults configures the default values for configuration
func setDefaults() {
	// Main config defaults
	viper.SetDefault("api_server_port", 8081)
	viper.SetDefault("mcp_server_url", "http://localhost:8080/events")
	viper.SetDefault("mcp_server_port", 8080)
	viper.SetDefault("tools_enabled", []string{"get_weather"})

	// Tracing config defaults
	viper.SetDefault("tracing.enabled", false)
	viper.SetDefault("tracing.service_name", "service")
	viper.SetDefault("tracing.endpoint_address", "localhost:4317")
	viper.SetDefault("tracing.timeout", "5s")
	viper.SetDefault("tracing.sampling_rate", 1.0)
	viper.SetDefault("tracing.batch_timeout", "5s")
	viper.SetDefault("tracing.environment", "development")
	viper.SetDefault("tracing.version", "0.1.0")

	// Auth config defaults
	viper.SetDefault("auth.domain", "")
	viper.SetDefault("auth.client_id", "")
	viper.SetDefault("auth.client_secret", "")
	viper.SetDefault("auth.callback_url", "")
	viper.SetDefault("auth.token_ttl", "1h")
	viper.SetDefault("auth.audience", "")

	// Google config defaults
	viper.SetDefault("google.client_id", "")
	viper.SetDefault("google.client_secret", "")
	viper.SetDefault("google.redirect_url", "")
	viper.SetDefault("google.scopes", []string{})
	viper.SetDefault("google.state_cookie", "")
	viper.SetDefault("google.token_source_file", "")
	viper.SetDefault("google.enabled", false)
}
