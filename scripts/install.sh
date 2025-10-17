#!/bin/bash
set -e

echo "Installing Service Control Panel..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Go is not installed. Please install Go 1.19+ first."
    exit 1
fi

# Build the Go binary
echo "Building Go binary..."
mask build

# Create necessary directories
mkdir -p logs

# Copy environment file
cp configs/environments/sample.env configs/environments/local.env
echo "Please edit configs/environments/local.env with your settings"

# Install systemd service
mkdir -p ~/.config/systemd/user
cp configs/systemd/service-control.service ~/.config/systemd/user/
systemctl --user daemon-reload
systemctl --user enable service-control

# Install nginx config
sudo cp configs/nginx/sites/service-control.conf /etc/nginx/sites-available/
sudo ln -sf /etc/nginx/sites-available/service-control.conf /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

echo "Installation complete!"
echo "1. Edit environments/local.env with your admin credentials"
echo "2. Start with: systemctl --user start service-control"
echo "3. Access at: https://service-control.aevion.lan"
echo ""
echo "To update: git pull && make build && systemctl --user restart service-control"
