package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"golang.org/x/exp/slog"

	"github.com/mikrocloud/mikrocloud/internal/api"
	"github.com/mikrocloud/mikrocloud/internal/config"
	"github.com/mikrocloud/mikrocloud/internal/database"
	"github.com/mikrocloud/mikrocloud/internal/domain/proxy"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers"
	serversService "github.com/mikrocloud/mikrocloud/internal/domain/servers/service"
	proxyContainers "github.com/mikrocloud/mikrocloud/pkg/containers/proxy"
)

type Server struct {
	config     *config.Config
	db         *database.Database
	router     *chi.Mux
	staticFS   fs.FS
	tokenAuth  *jwtauth.JWTAuth
	traefikSvc *proxyContainers.TraefikService
}

func New(cfg *config.Config, staticFS fs.FS) *Server {
	db, err := database.New(cfg)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}

	// Initialize Chi router
	router := chi.NewRouter()

	// initialize auth middleware

	tokenAuth := jwtauth.New("HS256", []byte(cfg.Auth.JWTSecret), nil)

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))
	router.Use(jwtauth.Verifier(tokenAuth))

	// Custom logging middleware using slog
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				slog.Info("HTTP Request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.Status(),
					"duration", time.Since(start),
					"bytes", ww.BytesWritten(),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	})

	return &Server{
		config:    cfg,
		db:        db,
		router:    router,
		staticFS:  staticFS,
		tokenAuth: tokenAuth,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Setup routes and get Traefik service
	if err := s.setupAPIRoutes(ctx); err != nil {
		return fmt.Errorf("failed to setup API routes: %w", err)
	}
	s.setupStaticRoutes()

	// Initialize default control plane server
	if err := s.initializeControlPlaneServer(ctx); err != nil {
		slog.Warn("Failed to initialize control plane server", "error", err)
		// Don't fail the server if this fails, just log the warning
	}

	// Start Traefik service if available
	if s.traefikSvc != nil {
		slog.Info("Starting Traefik reverse proxy service")

		// Create default global configuration for Traefik
		globalConfig := proxy.NewTraefikGlobalConfig()

		if err := s.traefikSvc.Start(ctx, globalConfig); err != nil {
			slog.Error("Failed to start Traefik service", "error", err)
			// Don't fail the server if Traefik fails to start
			// Just log the error and continue
		} else {
			slog.Info("Traefik service started successfully")
		}
	}

	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)

	server := &http.Server{
		Addr:    addr,
		Handler: s.router,
	}

	slog.Info("Starting Mikrocloud server", "address", addr)

	// Start server in a goroutine
	serverChan := make(chan error, 1)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverChan <- err
		}
	}()

	// Wait for interrupt signal or server error
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverChan:
		return fmt.Errorf("server error: %w", err)
	case sig := <-sigChan:
		slog.Info("Received shutdown signal", "signal", sig)

		// Graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop Traefik service first
		if s.traefikSvc != nil {
			slog.Info("Stopping Traefik service")
			if err := s.traefikSvc.Stop(shutdownCtx); err != nil {
				slog.Error("Error stopping Traefik service", "error", err)
			}
		}

		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		slog.Info("Server shutdown complete")
		return nil
	}
}

func (s *Server) setupAPIRoutes(ctx context.Context) error {
	var traefikSvc *proxyContainers.TraefikService
	var err error

	s.router.Route("/api", func(r chi.Router) {
		traefikSvc, _, err = api.SetupRoutes(r, s.db, s.config, s.tokenAuth, ctx)
	})

	if err != nil {
		return err
	}

	s.traefikSvc = traefikSvc
	return nil
}

func (s *Server) setupStaticRoutes() {
	frontendFS := s.staticFS
	if frontendFS == nil {
		slog.Error("No frontend assets available")
		s.setupPlaceholderRoutes()
		return
	}

	// Health check endpoint (always available)
	s.router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{
			"status": "ok",
			"service": "mikrocloud",
			"version": "0.1.0",
			"timestamp": "%s"
		}`, time.Now().UTC().Format(time.RFC3339))
	})

	// Single catch-all handler for everything non-API
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// API requests get JSON 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"API endpoint not found"}`)
			return
		}

		// Try to serve the file from the embedded FS first
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		file, err := frontendFS.Open(path)
		if err == nil {
			defer file.Close()

			// Set appropriate content type
			s.setContentType(w, path)

			// Set caching based on file type
			if strings.Contains(path, "/assets/") || strings.Contains(path, "/_app/") {
				w.Header().Set("Cache-Control", "max-age=31536000, immutable") // 1 year for versioned assets
			} else if strings.HasSuffix(path, ".html") {
				w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate") // No cache for HTML
			} else {
				w.Header().Set("Cache-Control", "max-age=86400") // 1 day for other static files
			}

			// Use ServeContent for better performance
			if stat, err := file.Stat(); err == nil {
				http.ServeContent(w, r, path, stat.ModTime(), file.(io.ReadSeeker))
			} else {
				io.Copy(w, file)
			}
			return
		}

		// File not found, serve index.html for SPA routing
		s.serveIndexHTML(frontendFS)(w, r)
	})
}

func (s *Server) setContentType(w http.ResponseWriter, path string) {
	switch {
	case strings.HasSuffix(path, ".html"):
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case strings.HasSuffix(path, ".css"):
		w.Header().Set("Content-Type", "text/css")
	case strings.HasSuffix(path, ".js"):
		w.Header().Set("Content-Type", "application/javascript")
	case strings.HasSuffix(path, ".json"):
		w.Header().Set("Content-Type", "application/json")
	case strings.HasSuffix(path, ".svg"):
		w.Header().Set("Content-Type", "image/svg+xml")
	case strings.HasSuffix(path, ".ico"):
		w.Header().Set("Content-Type", "image/x-icon")
	case strings.HasSuffix(path, ".png"):
		w.Header().Set("Content-Type", "image/png")
	case strings.HasSuffix(path, ".jpg"), strings.HasSuffix(path, ".jpeg"):
		w.Header().Set("Content-Type", "image/jpeg")
	case strings.HasSuffix(path, ".gif"):
		w.Header().Set("Content-Type", "image/gif")
	case strings.HasSuffix(path, ".woff"), strings.HasSuffix(path, ".woff2"):
		w.Header().Set("Content-Type", "font/woff")
	case strings.HasSuffix(path, ".ttf"):
		w.Header().Set("Content-Type", "font/ttf")
	}
}

// serveIndexHTML returns a handler that always serves the index.html file
func (s *Server) serveIndexHTML(frontendFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexFile, err := frontendFS.Open("index.html")
		if err != nil {
			slog.Error("Failed to open index.html", "error", err)
			http.Error(w, "Frontend not available", http.StatusInternalServerError)
			return
		}
		defer indexFile.Close()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")

		// Read the file content and serve it
		content, err := io.ReadAll(indexFile)
		if err != nil {
			slog.Error("Failed to read index.html", "error", err)
			http.Error(w, "Frontend not available", http.StatusInternalServerError)
			return
		}

		w.Write(content)
	}
}

func (s *Server) setupPlaceholderRoutes() {
	// Frontend placeholder (fallback when assets aren't available)
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Mikrocloud</title>
			<style>
				body { font-family: Arial, sans-serif; margin: 2rem; }
				.container { max-width: 800px; margin: 0 auto; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>🐳 Mikrocloud</h1>
				<p>Container management platform</p>
				<p><strong>Note:</strong> Frontend assets not built yet. Run <code>make build-web</code> to build and embed the frontend.</p>
				<p><a href="/docs">📖 API Documentation</a></p>
				<p><a href="/health">💚 Health Check</a></p>
			</div>
		</body>
		</html>`)
	})

	// Catch-all for SPA routing
	s.router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// If it's an API request, return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, "API endpoint not found", http.StatusNotFound)
			return
		}

		// Otherwise serve placeholder
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head><title>Mikrocloud - Page Not Found</title></head>
		<body>
			<h1>Mikrocloud - Page Not Found</h1>
			<p>The frontend hasn't been built yet. Please run <code>make build-web</code> first.</p>
			<p><a href="/">← Back to Home</a></p>
		</body>
		</html>`)
	})
}

func (s *Server) initializeControlPlaneServer(ctx context.Context) error {
	serversSvc := serversService.NewServersService(s.db.ServersRepository)

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}

	existingServer, err := serversSvc.GetServerByHostname(hostname)
	if err == nil && existingServer != nil {
		slog.Info("Control plane server already initialized", "hostname", hostname, "server_id", existingServer.ID())
		return nil
	}

	// Get the first organization from the database
	var orgID string
	err = s.db.MainDB().DB().QueryRowContext(ctx, "SELECT id FROM organizations LIMIT 1").Scan(&orgID)
	if err != nil {
		return fmt.Errorf("no organization found in database: %w", err)
	}
	defaultOrgID := uuid.MustParse(orgID)

	cpuCores := runtime.NumCPU()
	memoryMB := getSystemMemoryMB()
	diskGB := getDiskSpaceMB() / 1024

	osInfo := runtime.GOOS
	osVersion := runtime.Version()

	createdServer, err := serversSvc.CreateServer(
		"Control Plane - "+hostname,
		hostname,
		"127.0.0.1",
		s.config.Server.Port,
		servers.ServerTypeControlPlane,
		defaultOrgID,
	)
	if err != nil {
		return fmt.Errorf("failed to create control plane server: %w", err)
	}

	createdServer.UpdateDescription("Default control plane server for mikrocloud")
	createdServer.UpdateSpecs(&cpuCores, &memoryMB, &diskGB, &osInfo, &osVersion)
	createdServer.AddTag("control-plane")
	createdServer.AddTag("default")

	if err := serversSvc.UpdateServer(createdServer); err != nil {
		slog.Warn("Failed to update control plane server specs", "error", err)
	}

	slog.Info("Control plane server initialized",
		"server_id", createdServer.ID(),
		"hostname", hostname,
		"cpu_cores", cpuCores,
		"memory_mb", memoryMB,
	)

	return nil
}

func getSystemMemoryMB() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return int(m.Sys / 1024 / 1024)
}

func getDiskSpaceMB() int {
	return 100000
}
