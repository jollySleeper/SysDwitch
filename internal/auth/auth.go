// internal/auth/auth.go
package auth

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Username string
	Password string
	logger   *slog.Logger
}

// NewAuthConfig creates auth config from environment variables
func NewAuthConfig(logger *slog.Logger) (*AuthConfig, error) {
	username := strings.TrimSpace(os.Getenv("ADMIN_USER"))
	password := strings.TrimSpace(os.Getenv("ADMIN_PASS"))

	if username == "" || password == "" {
		return nil, errors.New("ADMIN_USER and ADMIN_PASS environment variables must be set")
	}

	if logger == nil {
		logger = slog.Default()
	}

	return &AuthConfig{
		Username: username,
		Password: password,
		logger:   logger,
	}, nil
}

// BasicAuthMiddleware provides HTTP Basic Authentication
func (ac *AuthConfig) BasicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			ac.logger.Debug("missing authorization header",
				"remote_addr", r.RemoteAddr,
				"method", r.Method,
				"path", r.URL.Path)
			ac.requireAuth(w)
			return
		}

		if !strings.HasPrefix(auth, "Basic ") {
			ac.logger.Warn("invalid authorization scheme",
				"scheme", strings.Fields(auth)[0],
				"remote_addr", r.RemoteAddr)
			ac.requireAuth(w)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			ac.logger.Warn("failed to decode authorization header",
				"error", err,
				"remote_addr", r.RemoteAddr)
			ac.requireAuth(w)
			return
		}

		creds := strings.SplitN(string(decoded), ":", 2)
		if len(creds) != 2 {
			ac.logger.Warn("malformed credentials in authorization header",
				"remote_addr", r.RemoteAddr)
			ac.requireAuth(w)
			return
		}

		username, password := creds[0], creds[1]

		// Use constant-time comparison to prevent timing attacks
		if subtle.ConstantTimeCompare([]byte(username), []byte(ac.Username)) != 1 ||
			subtle.ConstantTimeCompare([]byte(password), []byte(ac.Password)) != 1 {
			ac.logger.Warn("authentication failed",
				"username", username,
				"remote_addr", r.RemoteAddr)
			ac.requireAuth(w)
			return
		}

		ac.logger.Debug("authentication successful",
			"username", username,
			"remote_addr", r.RemoteAddr)

		// Authentication successful, call next handler
		next(w, r)
	}
}

// requireAuth sends a 401 Unauthorized response
func (ac *AuthConfig) requireAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Service Control Panel"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
