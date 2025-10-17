// cmd/service-control/main.go
package main

import (
	"context"
	"errors"
	"flag"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"service-control-panel/internal/auth"
	"service-control-panel/internal/handlers"
	"service-control-panel/internal/service"
	"service-control-panel/web"
)

// AppConfig holds application configuration
type AppConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	AllowedServices []string      `json:"allowed_services"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	ServiceManager  *service.ServiceManager
	AuthConfig      *auth.AuthConfig
}

// loadConfig loads configuration from environment variables and flags
func loadConfig() (*AppConfig, error) {
	var config AppConfig

	// Command line flags
	flag.StringVar(&config.Host, "host", getEnvOrDefault("HOST", "127.0.0.1"), "server host")
	flag.IntVar(&config.Port, "port", getEnvIntOrDefault("PORT", 8081), "server port")

	// Parse flags
	flag.Parse()

	// Get allowed services from environment
	allowedServicesStr := getEnvOrDefault("ALLOWED_SERVICES", "calibre.service,jellyfin.service,navidrome.service")
	config.AllowedServices = strings.Split(allowedServicesStr, ",")
	for i, s := range config.AllowedServices {
		config.AllowedServices[i] = strings.TrimSpace(s)
		if config.AllowedServices[i] == "" {
			return nil, errors.New("empty service name in ALLOWED_SERVICES")
		}
	}

	// HTTP timeouts
	config.ReadTimeout = 15 * time.Second
	config.WriteTimeout = 15 * time.Second

	// Validate configuration
	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New("invalid port number")
	}

	return &config, nil
}

// Helper functions for environment variable handling
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize components
	authConfig, err := auth.NewAuthConfig(logger)
	if err != nil {
		logger.Error("failed to initialize auth config", "error", err)
		os.Exit(1)
	}

	serviceManager := service.NewServiceManager(config.AllowedServices, logger)

	// Parse templates from embedded files
	templates, err := template.New("").Funcs(template.FuncMap{
		"trimSuffix": strings.TrimSuffix,
	}).ParseFS(web.TemplatesFS, "templates/index.html")
	if err != nil {
		logger.Error("failed to parse embedded templates", "error", err)
		os.Exit(1)
	}

	// Store references in config for use in handlers
	config.AuthConfig = authConfig
	config.ServiceManager = serviceManager

	// Create handler instance
	handler := handlers.NewHandler(logger, serviceManager, authConfig, templates)

	// Create HTTP server
	mux := http.NewServeMux()

	// Dashboard route
	mux.HandleFunc("/", authConfig.BasicAuthMiddleware(handler.Dashboard))

	// API routes for service control
	mux.HandleFunc("/api/services/", authConfig.BasicAuthMiddleware(handler.ServiceControl))

	// API status route
	mux.HandleFunc("/api/services/status", authConfig.BasicAuthMiddleware(handler.ServiceStatus))

	// Static files from embedded FS
	staticFS, err := fs.Sub(web.StaticFS, "static")
	if err != nil {
		logger.Error("failed to create static file subsystem", "error", err)
		os.Exit(1)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Apply middleware chain
	muxWithMiddleware := panicRecoveryMiddleware(logger)(
		requestLoggingMiddleware(logger)(
			securityHeadersMiddleware(mux)))

	// Configure HTTP server with timeouts and limits
	server := &http.Server{
		Addr:         config.Host + ":" + strconv.Itoa(config.Port),
		Handler:      muxWithMiddleware,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  60 * time.Second,
		// Limit request body size to prevent DoS
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Channel to listen for interrupt signals
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		logger.Info("starting Service Control Panel",
			"address", server.Addr,
			"allowed_services", config.AllowedServices)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	<-done
	logger.Info("received shutdown signal, shutting down gracefully...")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown complete")
}

// panicRecoveryMiddleware recovers from panics and logs them
func panicRecoveryMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered in HTTP handler",
						"panic", err,
						"url", r.URL.Path,
						"method", r.Method,
						"remote_addr", r.RemoteAddr)

					// Return 500 Internal Server Error
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// requestLoggingMiddleware logs all HTTP requests
func requestLoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			logger.Info("HTTP request",
				"method", r.Method,
				"url", r.URL.Path,
				"status", wrapper.statusCode,
				"duration", time.Since(start),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.Header.Get("User-Agent"))
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// securityHeadersMiddleware adds security headers to all responses
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// XSS protection (legacy, but still useful)
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer policy for privacy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Content Security Policy for additional protection
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; style-src 'self' 'unsafe-inline' https://cdn.tailwindcss.com; img-src 'self' data:;")

		// HSTS (HTTP Strict Transport Security) - only if using HTTPS
		// w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		next.ServeHTTP(w, r)
	})
}
