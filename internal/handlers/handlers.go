// internal/handlers/handlers.go
package handlers

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"
	"strings"

	"sysdwitch/internal/auth"
	"sysdwitch/internal/service"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	logger         *slog.Logger
	serviceManager *service.ServiceManager
	authConfig     *auth.AuthConfig
	templates      *template.Template
}

// NewHandler creates a new handler instance
func NewHandler(logger *slog.Logger, serviceManager *service.ServiceManager, authConfig *auth.AuthConfig, templates *template.Template) *Handler {
	return &Handler{
		logger:         logger,
		serviceManager: serviceManager,
		authConfig:     authConfig,
		templates:      templates,
	}
}

// Dashboard renders the main dashboard page
func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Warn("invalid method for dashboard",
			"method", r.Method, "remote_addr", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	services := h.serviceManager.GetAllServicesStatus(ctx)
	data := struct {
		Services []service.ServiceStatus
	}{
		Services: services,
	}

	if err := h.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		h.logger.Error("template execution error",
			"error", err, "template", "index.html", "remote_addr", r.RemoteAddr)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// ServiceControl handles service start/stop operations
func (h *Handler) ServiceControl(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract service name from URL path
	path := strings.TrimPrefix(r.URL.Path, "/api/services/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		h.logger.Warn("invalid API path format",
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
			h.logger.Warn("invalid method for service start",
				"method", r.Method, "service", serviceName, "remote_addr", r.RemoteAddr)
			response = APIResponse{Success: false, Error: "Method not allowed"}
			break
		}
		service := h.serviceManager.StartService(ctx, serviceName)
		response = APIResponse{Success: true, Service: &service}
		h.logger.Info("service start requested",
			"service", serviceName, "status", service.Status, "remote_addr", r.RemoteAddr)

	case "stop":
		if r.Method != http.MethodPost {
			h.logger.Warn("invalid method for service stop",
				"method", r.Method, "service", serviceName, "remote_addr", r.RemoteAddr)
			response = APIResponse{Success: false, Error: "Method not allowed"}
			break
		}
		service := h.serviceManager.StopService(ctx, serviceName)
		response = APIResponse{Success: true, Service: &service}
		h.logger.Info("service stop requested",
			"service", serviceName, "status", service.Status, "remote_addr", r.RemoteAddr)

	default:
		h.logger.Warn("invalid action requested",
			"action", action, "service", serviceName, "remote_addr", r.RemoteAddr)
		response = APIResponse{Success: false, Error: "Invalid action. Supported: start, stop"}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode JSON response",
			"error", err, "remote_addr", r.RemoteAddr)
	}
}

// ServiceStatus returns the status of all services
func (h *Handler) ServiceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.logger.Warn("invalid method for status endpoint",
			"method", r.Method, "remote_addr", r.RemoteAddr)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()
	services := h.serviceManager.GetAllServicesStatus(ctx)
	response := APIResponse{Success: true, Services: services}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode JSON response for status",
			"error", err, "remote_addr", r.RemoteAddr)
	}
}

// APIResponse represents API response structure
type APIResponse struct {
	Success  bool                    `json:"success"`
	Service  *service.ServiceStatus  `json:"service,omitempty"`
	Services []service.ServiceStatus `json:"services,omitempty"`
	Error    string                  `json:"error,omitempty"`
}
