# sysdwitch

**SystemD Switch** - A secure web-based control panel for managing systemd user services with a modern, responsive interface.

> ⚠️ **Security Notice**
> This application provides authenticated access to systemd service management. Ensure proper network security (HTTPS, firewall rules) and use strong passwords. Only expose to trusted networks.

## 📖 Table of Contents

- [✨ Features](#features)
- [🚀 Installation](#installation)
- [🛠️ Usage](#usage)
- [⚙️ Configuration](#configuration)
- [🔧 Development](#development)
- [📚 Documentation](#documentation)
- [🐛 Bugs or Requests](#bugs-or-requests)
- [🤝 Contributing](#contributing)
- [📄 License](#license)
- [🙏 Acknowledgments](#acknowledgments)

## ✨ Features

- **🔒 Secure Authentication**: HTTP Basic Auth with constant-time password comparison
- **🏗️ Single Binary**: Embedded HTML/CSS/JS assets for easy deployment
- **📊 Structured Logging**: Comprehensive logging with slog (Go 1.25+)
- **⚡ High Performance**: Optimized for low latency with embedded assets
- **🔄 Service Management**: Start/stop systemd user services with real-time status
- **📱 Responsive UI**: Modern TailwindCSS interface that works on all devices
- **🛡️ Security Headers**: CSP, XSS protection, and other security measures
- **🔥 Rate Limiting**: Built-in rate limiting to prevent abuse
- **📈 Graceful Shutdown**: Proper cleanup and signal handling
- **🐳 Container Ready**: Multi-stage Docker builds with security best practices
- **🚦 Health Monitoring**: Built-in health checks and service status monitoring
- **🔧 Configuration**: Environment-based configuration with validation

## 🚀 Installation

### Prerequisites
- Go 1.25 or later
- Linux with systemd
- Internet connection (for TailwindCSS CDN)

### Quick Start

#### Using Go directly:
```bash
# Clone or download the project
cd sysdwitch

# Run the server
go run ./cmd/sysdwitch

# Or build and run
go build -o sysdwitch ./cmd/sysdwitch
```

The server starts on port 8081 by default.

#### Using Docker:
```bash
# Build the Docker image
docker build -t sysdwitch .

# Run the container
docker run -p 8081:8081 \
  -e ADMIN_USER=admin \
  -e ADMIN_PASS=secure_password \
  -e ALLOWED_SERVICES=calibre,jellyfin \
  sysdwitch
```

#### Production Deployment:
```bash
# Install using the provided script
./scripts/install.sh

# Or manually:
# 1. Build the binary
go build -o sysdwitch ./cmd/sysdwitch

# 2. Copy systemd service
cp init/sysdwitch.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable sysdwitch

# 3. Configure environment
cp configs/environments/sample.env configs/environments/local.env
# Edit local.env with your settings

# 4. Start the service
systemctl --user start sysdwitch
```

## 🛠️ Usage

### URL Format
Access the control panel at: `http://localhost:8081`

### Supported Services
- **User Services**: Any systemd --user service in your whitelist
- **Service Actions**: Start, stop, and status monitoring
- **Real-time Updates**: Automatic status refresh every 30 seconds

### Examples
```bash
# Check service status
curl -u admin:password http://localhost:8081/api/services/status

# Start a service
curl -u admin:password -X POST http://localhost:8081/api/services/calibre/start

# Stop a service
curl -u admin:password -X POST http://localhost:8081/api/services/jellyfin/stop
```

## ⚙️ Configuration

### Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| `ADMIN_USER` | *required* | Admin username for authentication |
| `ADMIN_PASS` | *required* | Admin password for authentication |
| `ALLOWED_SERVICES` | `calibre,jellyfin,navidrome` | Comma-separated service names |
| `HOST` | `127.0.0.1` | Server bind address |
| `PORT` | `8081` | Server port |
| `LOG_LEVEL` | `info` | Logging level (debug, info, warn, error) |

### Examples
```bash
# Basic configuration
ADMIN_USER=admin ADMIN_PASS=secure_password go run ./cmd/sysdwitch

# Development with debug logging
LOG_LEVEL=debug ADMIN_USER=admin ADMIN_PASS=test go run ./cmd/sysdwitch

# Custom services and port
PORT=3000 ALLOWED_SERVICES=nginx,mysql,redis ADMIN_USER=admin ADMIN_PASS=secure go run ./cmd/sysdwitch
```

### Docker Configuration
```bash
docker run -p 8081:8081 \
  -e ADMIN_USER=myuser \
  -e ADMIN_PASS=mypass \
  -e ALLOWED_SERVICES=service1,service2 \
  -e LOG_LEVEL=debug \
  sysdwitch
```

## 🔧 Development

### Quick Setup
```bash
# Install dependencies
go mod download

# Run in development mode
LOG_LEVEL=debug go run ./cmd/sysdwitch

# Build for production
go build -o sysdwitch ./cmd/sysdwitch

# Run tests
go test ./...
```

### Project Structure
```
sysdwitch/
├── cmd/sysdwitch/          # Application entry point
│   └── main.go            # Main function and startup logic
├── internal/              # Private application code
│   ├── auth/              # Authentication middleware
│   ├── handlers/          # HTTP request handlers
│   └── service/           # Service management logic
├── web/                   # Embedded web assets
│   ├── static/           # CSS, JS, images
│   └── templates/        # HTML templates
├── configs/               # Configuration files
│   ├── environments/     # Environment configurations
│   └── nginx/            # Nginx proxy config
├── init/                  # System init configs
├── scripts/               # Build and deployment scripts
└── maskfile.md           # Task automation
```

### Key Technologies
- **Go 1.25**: Latest language features and optimizations
- **Structured Logging**: `log/slog` package for observability
- **Embedded Assets**: `//go:embed` for single binary deployment
- **HTTP Security**: Security headers and rate limiting
- **Systemd Integration**: Direct systemctl --user commands

### Build & Test
```bash
# Build binary
go build -o sysdwitch ./cmd/sysdwitch

# Run tests with coverage
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint code
golangci-lint run

# Format code
go fmt ./...
```

## 📚 Documentation

Comprehensive documentation is available in the project files:

### 🏗️ **Architecture & Design**
- [maskfile.md](./maskfile.md) - Task automation definitions

### 🔧 **Core Components**
- **Service Manager**: Systemd user service control with validation
- **Authentication**: HTTP Basic Auth with secure password checking
- **Web Interface**: Embedded HTML/CSS/JS with TailwindCSS
- **Security**: Rate limiting, security headers, input validation

### 📋 **API Reference**
- `GET /` - Main dashboard (requires auth)
- `GET /api/services/status` - Get all service statuses
- `POST /api/services/{name}/start` - Start a service
- `POST /api/services/{name}/stop` - Stop a service
- `GET /static/*` - Static assets (CSS, JS, images)

### 🚀 **Quick Access**
```bash
# Open main documentation
open IMPLEMENTATION.md

# View project structure
tree -I 'vendor|node_modules'

# Check service status
systemctl --user status sysdwitch
```

## 🐛 Bugs or Requests

### Troubleshooting

#### Common Issues
1. **"Permission denied" errors**
   - Ensure the user has permission to run systemctl --user commands
   - Check systemd user service permissions

2. **"Service not allowed" errors**
   - Verify the service name is in ALLOWED_SERVICES
   - Check service name format (without .service extension in config)

3. **Authentication failures**
   - Verify ADMIN_USER and ADMIN_PASS environment variables
   - Check for special characters in credentials

4. **Port already in use**
   - Change PORT environment variable
   - Kill existing processes on the port

### Reporting Issues
Please report bugs or request features by opening an [issue](https://github.com/jollySleeper/sysdwitch/issues/new) with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- System information (Go version, OS, systemd version)

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go 1.25 best practices
- Add tests for new functionality
- Update documentation for API changes
- Use structured logging with appropriate levels
- Maintain security best practices

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

This is free and open-source software. You are free to use, modify, and distribute it for personal and commercial use.

## 🙏 Acknowledgments

This project was built following Go best practices and security guidelines. Special thanks to:

- The Go team for an excellent programming language and standard library
- The systemd project for reliable service management
- The TailwindCSS team for beautiful, responsive styling
- Open source community for security research and best practices
