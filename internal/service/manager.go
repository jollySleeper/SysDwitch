// internal/service/manager.go
package service

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ServiceStatus represents the status of a systemd service
type ServiceStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Active bool   `json:"active"`
}

// ServiceManager handles systemd service operations
type ServiceManager struct {
	allowedServices map[string]bool
	logger          *slog.Logger
	mu              sync.RWMutex
}

// NewServiceManager creates a new service manager with allowed services
func NewServiceManager(allowedServices []string, logger *slog.Logger) *ServiceManager {
	allowed := make(map[string]bool)
	for _, service := range allowedServices {
		if !strings.HasSuffix(service, ".service") {
			service += ".service"
		}
		allowed[service] = true
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &ServiceManager{
		allowedServices: allowed,
		logger:          logger,
	}
}

// validateService checks if a service is in the allowed list
func (sm *ServiceManager) validateService(serviceName string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.allowedServices[serviceName]
}

// runSystemctl executes systemctl commands with timeout and context
func (sm *ServiceManager) runSystemctl(ctx context.Context, args ...string) (string, error) {
	// Create context with timeout for systemctl operations
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, "systemctl", append([]string{"--user"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		sm.logger.Error("systemctl command failed",
			"args", args,
			"error", err,
			"stderr", stderr.String())
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

// GetServiceStatus gets the status of a systemd user service
func (sm *ServiceManager) GetServiceStatus(ctx context.Context, serviceName string) ServiceStatus {
	if !sm.validateService(serviceName) {
		sm.logger.Warn("attempted to check status of non-allowed service",
			"service", serviceName)
		return ServiceStatus{Name: serviceName, Status: "not_allowed", Active: false}
	}

	status, err := sm.runSystemctl(ctx, "is-active", serviceName)
	if err != nil {
		sm.logger.Error("failed to get status for service",
			"service", serviceName,
			"error", err)
		return ServiceStatus{Name: serviceName, Status: "error", Active: false}
	}

	return ServiceStatus{
		Name:   serviceName,
		Status: status,
		Active: status == "active",
	}
}

// StartService starts a systemd user service
func (sm *ServiceManager) StartService(ctx context.Context, serviceName string) ServiceStatus {
	if !sm.validateService(serviceName) {
		sm.logger.Warn("attempted to start non-allowed service",
			"service", serviceName)
		return ServiceStatus{Name: serviceName, Status: "not_allowed", Active: false}
	}

	_, err := sm.runSystemctl(ctx, "start", serviceName)
	if err != nil {
		sm.logger.Error("failed to start service",
			"service", serviceName,
			"error", err)
		return ServiceStatus{Name: serviceName, Status: "error", Active: false}
	}

	return sm.GetServiceStatus(ctx, serviceName)
}

// StopService stops a systemd user service
func (sm *ServiceManager) StopService(ctx context.Context, serviceName string) ServiceStatus {
	if !sm.validateService(serviceName) {
		sm.logger.Warn("attempted to stop non-allowed service",
			"service", serviceName)
		return ServiceStatus{Name: serviceName, Status: "not_allowed", Active: false}
	}

	_, err := sm.runSystemctl(ctx, "stop", serviceName)
	if err != nil {
		sm.logger.Error("failed to stop service",
			"service", serviceName,
			"error", err)
		return ServiceStatus{Name: serviceName, Status: "error", Active: false}
	}

	return sm.GetServiceStatus(ctx, serviceName)
}

// GetAllServicesStatus gets status of all configured services
func (sm *ServiceManager) GetAllServicesStatus(ctx context.Context) []ServiceStatus {
	sm.mu.RLock()
	services := make([]string, 0, len(sm.allowedServices))
	for service := range sm.allowedServices {
		services = append(services, service)
	}
	sm.mu.RUnlock()

	results := make([]ServiceStatus, len(services))
	for i, service := range services {
		results[i] = sm.GetServiceStatus(ctx, service)
	}

	return results
}
