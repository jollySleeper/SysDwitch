// cmd/service-control/main.go
package main

import (
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"service-control-panel/internal/auth"
	"service-control-panel/internal/service"
)

//go:embed web/templates/*
//go:embed web/static/*
var embeddedFiles embed.FS

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

// APIResponse represents API response structure
type APIResponse struct {
	Success  bool                    `json:"success"`
	Service  *service.ServiceStatus  `json:"service,omitempty"`
	Services []service.ServiceStatus `json:"services,omitempty"`
	Error    string                  `json:"error,omitempty"`
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
	templates, err := template.ParseFS(embeddedFiles, "web/templates/*.html")
	if err != nil {
		logger.Error("failed to parse embedded templates", "error", err)
		os.Exit(1)
	}

	// Store references in config for use in handlers
	config.AuthConfig = authConfig
	config.ServiceManager = serviceManager

	// Create HTTP server
	mux := http.NewServeMux()

	// Dashboard route
	mux.HandleFunc("/", authConfig.BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Warn("invalid method for dashboard",
				"method", r.Method, "remote_addr", r.RemoteAddr)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		services := serviceManager.GetAllServicesStatus(ctx)
		data := struct {
			Services []service.ServiceStatus
		}{
			Services: services,
		}

		if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
			logger.Error("template execution error",
				"error", err, "template", "index.html", "remote_addr", r.RemoteAddr)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}))

	// API routes for service control
	mux.HandleFunc("/api/services/", authConfig.BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract service name from URL path
		path := strings.TrimPrefix(r.URL.Path, "/api/services/")
		parts := strings.Split(path, "/")
		if len(parts) < 2 {
			logger.Warn("invalid API path format",
				"path", r.URL.Path, "remote_addr", r.RemoteAddr)
			http.Error(w, `{"error":"Invalid path format. Expected /api/services/{name}/{action}"}`, http.StatusBadRequest)
			return
		}

		serviceName := parts[0]
		if !strings.HasSuffix(serviceName, ".service") {
			serviceName += ".service"
		}
		action := parts[1]

		ctx := r.Context()
		var response APIResponse

		switch action {
		case "start":
			if r.Method != http.MethodPost {
				logger.Warn("invalid method for service start",
					"method", r.Method, "service", serviceName, "remote_addr", r.RemoteAddr)
				response = APIResponse{Success: false, Error: "Method not allowed"}
				break
			}
			service := serviceManager.StartService(ctx, serviceName)
			response = APIResponse{Success: true, Service: &service}
			logger.Info("service start requested",
				"service", serviceName, "status", service.Status, "remote_addr", r.RemoteAddr)

		case "stop":
			if r.Method != http.MethodPost {
				logger.Warn("invalid method for service stop",
					"method", r.Method, "service", serviceName, "remote_addr", r.RemoteAddr)
				response = APIResponse{Success: false, Error: "Method not allowed"}
				break
			}
			service := serviceManager.StopService(ctx, serviceName)
			response = APIResponse{Success: true, Service: &service}
			logger.Info("service stop requested",
				"service", serviceName, "status", service.Status, "remote_addr", r.RemoteAddr)

		default:
			logger.Warn("invalid action requested",
				"action", action, "service", serviceName, "remote_addr", r.RemoteAddr)
			response = APIResponse{Success: false, Error: "Invalid action. Supported: start, stop"}
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode JSON response",
				"error", err, "remote_addr", r.RemoteAddr)
		}
	}))

	// API status route
	mux.HandleFunc("/api/services/status", authConfig.BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			logger.Warn("invalid method for status endpoint",
				"method", r.Method, "remote_addr", r.RemoteAddr)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()
		services := serviceManager.GetAllServicesStatus(ctx)
		response := APIResponse{Success: true, Services: services}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.Error("failed to encode JSON response for status",
				"error", err, "remote_addr", r.RemoteAddr)
		}
	}))

	// Static files from embedded FS
	staticFS, err := fs.Sub(embeddedFiles, "web/static")
	if err != nil {
		logger.Error("failed to create static file subsystem", "error", err)
		os.Exit(1)
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Security headers middleware
	secureMux := securityHeadersMiddleware(mux)

	// Configure HTTP server with timeouts and limits
	server := &http.Server{
		Addr:         config.Host + ":" + strconv.Itoa(config.Port),
		Handler:      secureMux,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  60 * time.Second,
		// Limit request body size to prevent DoS
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Start server
	logger.Info("starting Service Control Panel",
		"address", server.Addr,
		"allowed_services", config.AllowedServices)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server failed to start", "error", err)
		os.Exit(1)
	}
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
