package tools

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// BaseToolConfig is the base configuration that all tools embed
type BaseToolConfig struct {
	Enabled bool `mapstructure:"enabled" yaml:"enabled" validate:"required"`
}

func (b BaseToolConfig) IsEnabled() bool {
	return b.Enabled
}

// PostgresConfig represents PostgreSQL database configuration
type PostgresConfig struct {
	BaseToolConfig `mapstructure:",squash"`
	Databases      []PostgresDatabase `mapstructure:"databases" yaml:"databases" validate:"required,dive"`
}

type PostgresDatabase struct {
	Name            string   `mapstructure:"name" yaml:"name" validate:"required"`
	Host            string   `mapstructure:"host" yaml:"host" validate:"required"`
	Username        string   `mapstructure:"username" yaml:"username" validate:"required"`
	Password        string   `mapstructure:"password" yaml:"password" validate:"required"`
	Port            int      `mapstructure:"port" yaml:"port" validate:"required,min=1,max=65535"`
	SSLMode         string   `mapstructure:"sslmode" yaml:"sslmode" validate:"required,oneof=disable enable verify-full verify-ca"`
	BlockedCommands []string `mapstructure:"blocked_commands" yaml:"blocked_commands" validate:"dive"`
}

// GithubBaseConfig represents base GitHub configuration
type GithubBaseConfig struct {
	BaseToolConfig `mapstructure:",squash"`
	Token          string `mapstructure:"token" yaml:"token" validate:"required"`
}

// FilesystemConfig represents filesystem tool configuration
type FilesystemConfig struct {
	BaseToolConfig   `mapstructure:",squash"`
	AllowedDirectory string   `mapstructure:"allowed_directory" yaml:"allowed_directory" validate:"required,dir"`
	BlockedPattern   []string `mapstructure:"blocked_pattern" yaml:"blocked_pattern" validate:"dive,min=1"`
}

// GitConfig represents git tool configuration
type GitConfig struct {
	BaseToolConfig  `mapstructure:",squash"`
	DefaultRepoPath string   `mapstructure:"default_repo_path" yaml:"default_repo_path" validate:"required,dir"`
	BlockedCommands []string `mapstructure:"blocked_commands" yaml:"blocked_commands" validate:"dive,min=1"`
}

// CurlConfig represents curl tool configuration
type CurlConfig struct {
	BaseToolConfig `mapstructure:",squash"`
	BlockedMethods []string `mapstructure:"blocked_methods" yaml:"blocked_methods" validate:"dive,oneof=GET POST PUT DELETE PATCH HEAD OPTIONS"`
}

// SimpleToolConfig represents tools with just enabled/disabled status
type SimpleToolConfig struct {
	BaseToolConfig `mapstructure:",squash"`
}

// GmailConfig represents Gmail tool configuration
type GmailConfig struct {
	BaseToolConfig `mapstructure:",squash"`
	UserID         string `mapstructure:"user_id" yaml:"user_id" validate:"required"`
	MaxResults     int64  `mapstructure:"max_results" yaml:"max_results" validate:"required,min=1"`
	SinceLastNDays int    `mapstructure:"since_last_n_days" yaml:"since_last_n_days" validate:"required,min=1"`
	Token          string `mapstructure:"token" yaml:"token" validate:"required"`
}

// WeatherConfig represents weather tool configuration
type WeatherConfig struct {
	BaseToolConfig `mapstructure:",squash"`
}

// ToolsConfig is the main configuration structure for all tools
type ToolsConfig struct {
	Weather          *WeatherConfig    `mapstructure:"get_wether" yaml:"get_wether"`
	Postgres         *PostgresConfig   `mapstructure:"postgres" yaml:"postgres"`
	GithubRepository *GithubBaseConfig `mapstructure:"github_repository" yaml:"github_repository"`
	GithubIssues     *GithubBaseConfig `mapstructure:"github_issues" yaml:"github_issues"`
	GithubPulls      *GithubBaseConfig `mapstructure:"github_pull_requests" yaml:"github_pull_requests"`
	GithubSearch     *GithubBaseConfig `mapstructure:"github_search" yaml:"github_search"`
	Filesystem       *FilesystemConfig `mapstructure:"filesystem" yaml:"filesystem"`
	Git              *GitConfig        `mapstructure:"git" yaml:"git"`
	Curl             *CurlConfig       `mapstructure:"curl" yaml:"curl"`
	Bash             *SimpleToolConfig `mapstructure:"bash" yaml:"bash"`
	Sed              *SimpleToolConfig `mapstructure:"sed" yaml:"sed"`
	Grep             *SimpleToolConfig `mapstructure:"grep" yaml:"grep"`
	Cat              *SimpleToolConfig `mapstructure:"cat" yaml:"cat"`
	Gmail            *GmailConfig      `mapstructure:"gmail" yaml:"gmail"`
	Docker           *SimpleToolConfig `mapstructure:"docker" yaml:"docker"`
}

// Validate implements validation for the entire configuration
func (t *ToolsConfig) Validate() error {
	validate := validator.New()

	if t.Postgres != nil && t.Postgres.IsEnabled() {
		if err := validate.Struct(t.Postgres); err != nil {
			return fmt.Errorf("postgres config validation failed: %w", err)
		}
	}

	// GitHub tools validation
	githubConfigs := map[string]*GithubBaseConfig{
		"github_repository":    t.GithubRepository,
		"github_issues":        t.GithubIssues,
		"github_pull_requests": t.GithubPulls,
		"github_search":        t.GithubSearch,
	}

	for name, config := range githubConfigs {
		if config != nil && config.IsEnabled() {
			if err := validate.Struct(config); err != nil {
				return fmt.Errorf("%s config validation failed: %w", name, err)
			}
		}
	}

	return nil
}

// SetDefaults sets the default values for tools configuration
func SetDefaults(v *viper.Viper) {
	// Set defaults for tools configuration
	v.SetDefault("tools.get_wether.enabled", false)
	v.SetDefault("tools.postgres.enabled", false)
	v.SetDefault("tools.github_repository.enabled", false)
	v.SetDefault("tools.github_issues.enabled", false)
	v.SetDefault("tools.github_pull_requests.enabled", false)
	v.SetDefault("tools.github_search.enabled", false)
	v.SetDefault("tools.filesystem.enabled", false)
	v.SetDefault("tools.git.enabled", false)
	v.SetDefault("tools.curl.enabled", false)
	v.SetDefault("tools.bash.enabled", false)
	v.SetDefault("tools.sed.enabled", false)
	v.SetDefault("tools.grep.enabled", false)
	v.SetDefault("tools.gmail.enabled", false)
	v.SetDefault("tools.docker.enabled", false)
	v.SetDefault("tools.cat.enabled", false)

	// Set other tool-specific defaults as needed
	v.SetDefault("tools.filesystem.allowed_directory", "/tmp")
	v.SetDefault("tools.git.default_repo_path", "/tmp")
	v.SetDefault("tools.curl.blocked_methods", []string{"POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"})
	v.SetDefault("tools.gmail.token", "")
}
