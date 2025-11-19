package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Analytics AnalyticsConfig `mapstructure:"analytics"`
	Queue     QueueConfig     `mapstructure:"queue"`
	Docker    DockerConfig    `mapstructure:"docker"`
	SSL       SSLConfig       `mapstructure:"ssl"`
	Auth      AuthConfig      `mapstructure:"auth"`
}

type ServerConfig struct {
	Host           string   `mapstructure:"host"`
	Port           int      `mapstructure:"port"`
	DataDir        string   `mapstructure:"data_dir"`
	LogLevel       string   `mapstructure:"log_level"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	PublicIP       string   `mapstructure:"public_ip"`
}

type DatabaseConfig struct {
	Type string `mapstructure:"type"` // "sqlite" or "postgres"
	URL  string `mapstructure:"url"`  // File path for SQLite, connection string for PostgreSQL
}

type AnalyticsConfig struct {
	Type string `mapstructure:"type"` // "duckdb" or "clickhouse"
	URL  string `mapstructure:"url"`  // File path for DuckDB, connection string for ClickHouse
}

type QueueConfig struct {
	Type string `mapstructure:"type"` // "dragonfly"
	URL  string `mapstructure:"url"`  // Connection string for Dragonfly
}

type DockerConfig struct {
	Runtime     string `mapstructure:"runtime"` // "docker" or "podman"
	SocketPath  string `mapstructure:"socket_path"`
	Rootless    bool   `mapstructure:"rootless"`
	BuildDir    string `mapstructure:"build_dir"`    // Directory for build workspaces
	NetworkMode string `mapstructure:"network_mode"` // Network mode for proxy: "bridge" or "host"
}

type SSLConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	ACMEEmail string `mapstructure:"acme_email"`
	Staging   bool   `mapstructure:"staging"`
	CertsDir  string `mapstructure:"certs_dir"`
}

type AuthConfig struct {
	JWTSecret string `mapstructure:"jwt_secret"`
	Enabled   bool   `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	var cfg Config

	// Set defaults
	setDefaults()

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand environment variables in paths
	cfg.Server.DataDir = expandEnvVars(cfg.Server.DataDir)
	cfg.Database.URL = expandEnvVars(cfg.Database.URL)
	cfg.Analytics.URL = expandEnvVars(cfg.Analytics.URL)
	cfg.Queue.URL = expandEnvVars(cfg.Queue.URL)
	cfg.SSL.CertsDir = expandEnvVars(cfg.SSL.CertsDir)
	cfg.Docker.BuildDir = expandEnvVars(cfg.Docker.BuildDir)

	return &cfg, nil
}

func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 3000)
	viper.SetDefault("server.data_dir", "${HOME}/.local/share/mikrocloud")
	viper.SetDefault("server.log_level", "info")
	viper.SetDefault("server.allowed_origins", []string{"*"})
	viper.SetDefault("server.public_ip", "")

	// Database defaults - SQLite database path
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.url", "${HOME}/.local/share/mikrocloud/mikrocloud.db")

	// Analytics defaults - DuckDB database path
	viper.SetDefault("analytics.type", "duckdb")
	viper.SetDefault("analytics.url", "${HOME}/.local/share/mikrocloud/analytics.duckdb")

	// Queue defaults - Dragonfly connection
	viper.SetDefault("queue.type", "dragonfly")
	viper.SetDefault("queue.url", "redis://localhost:6379/0")

	// Docker defaults
	viper.SetDefault("docker.runtime", "docker")
	viper.SetDefault("docker.socket_path", "/var/run/docker.sock")
	viper.SetDefault("docker.rootless", false)
	viper.SetDefault("docker.build_dir", "${HOME}/.local/share/mikrocloud/builds")
	viper.SetDefault("docker.network_mode", "bridge")

	// SSL defaults
	viper.SetDefault("ssl.enabled", false)
	viper.SetDefault("ssl.staging", true)
	viper.SetDefault("ssl.certs_dir", "${HOME}/.local/share/mikrocloud/certs")

	// Auth defaults
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.jwt_secret", "")
}

func expandEnvVars(path string) string {
	if strings.Contains(path, "${") {
		return os.ExpandEnv(path)
	}
	return path
}
