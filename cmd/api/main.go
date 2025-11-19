package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mikrocloud/mikrocloud/assets"
	"github.com/mikrocloud/mikrocloud/internal/config"
	"github.com/mikrocloud/mikrocloud/internal/server"
)

var (
	configFile string
	staticFS   fs.FS
	rootCmd    = &cobra.Command{
		Use:   "mikrocloud",
		Short: "Ultra-lightweight Platform as a Service (PaaS)",
		Long:  `Mikrocloud is a next-generation, multi-region Platform as a Service (PaaS) built for ultra-lightweight performance (<50MB memory usage) with enterprise features.`,
	}
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	// Set up static filesystem
	frontendFS, err := fs.Sub(assets.FrontendFS, "dist")
	if err != nil {
		slog.Error("Failed to get static filesystem", "error", err)
		os.Exit(1)
	}

	staticFS = frontendFS

	ctx := context.Background()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./mikrocloud.toml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")

	viper.BindPFlag("log_level", rootCmd.PersistentFlags().Lookup("log-level"))

	// Add subcommands
	// TODO: The migration should be automatic, no need for commmand
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(migrateCmd)
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("mikrocloud")
		viper.SetConfigType("toml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.config/mikrocloud")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "file", viper.ConfigFileUsed())
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Mikrocloud server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		srv := server.New(cfg, staticFS)

		return srv.Start(cmd.Context())
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Mikrocloud v%s\n", "0.1.0")
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations for all database systems",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Migrate main database
		if err := migrateMainDatabase(cfg); err != nil {
			return fmt.Errorf("failed to migrate main database: %w", err)
		}

		// Migrate analytics database
		if err := migrateAnalyticsDatabase(cfg); err != nil {
			return fmt.Errorf("failed to migrate analytics database: %w", err)
		}

		// Queue database doesn't typically need migrations for Redis/Dragonfly
		slog.Info("Queue database initialized", "type", cfg.Queue.Type, "url", cfg.Queue.URL)

		slog.Info("All database migrations completed successfully")
		return nil
	},
}

func migrateMainDatabase(cfg *config.Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Database.URL)
	if err := ensureDir(dbDir); err != nil {
		return fmt.Errorf("failed to create main database directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("failed to open main database: %w", err)
	}
	defer db.Close()

	// Set up goose for main database
	goose.SetDialect("sqlite3")

	// Run migrations from main migrations directory
	if err := goose.Up(db, "./migrations/main"); err != nil {
		return fmt.Errorf("failed to run main database migrations: %w", err)
	}

	slog.Info("Main database migrations completed successfully", "database", cfg.Database.URL)
	return nil
}

func migrateAnalyticsDatabase(cfg *config.Config) error {
	// Ensure database directory exists
	dbDir := filepath.Dir(cfg.Analytics.URL)
	if err := ensureDir(dbDir); err != nil {
		return fmt.Errorf("failed to create analytics database directory: %w", err)
	}

	// DuckDB schema is managed by the initSchema() method in duckdb.go
	// SQLite analytics would use goose migrations
	if cfg.Analytics.Type == "duckdb" {
		slog.Info("Analytics database schema managed by application code", "database", cfg.Analytics.URL)
		return nil
	}

	// For SQLite analytics, use goose migrations
	db, err := sql.Open("sqlite3", cfg.Analytics.URL)
	if err != nil {
		return fmt.Errorf("failed to open analytics database: %w", err)
	}
	defer db.Close()

	goose.SetDialect("sqlite3")

	if err := goose.Up(db, "./migrations/analytics"); err != nil {
		return fmt.Errorf("failed to run analytics database migrations: %w", err)
	}

	slog.Info("Analytics database migrations completed successfully", "database", cfg.Analytics.URL)
	return nil
}

func ensureDir(dir string) error {
	if dir == "." || dir == "/" {
		return nil
	}
	return os.MkdirAll(dir, 0o755)
}
