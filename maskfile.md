# Service Control Panel Tasks

## build

Build the application binary.

```bash
go build -o sysdwitch -v ./cmd/sysdwitch
```

## build-linux

Build for Linux (cross-compilation).

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sysdwitch-linux-amd64 -v ./cmd/sysdwitch
```

## build-all

Build for multiple platforms.

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sysdwitch-linux-amd64 ./cmd/sysdwitch
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o sysdwitch-darwin-amd64 ./cmd/sysdwitch
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o sysdwitch-windows-amd64.exe ./cmd/sysdwitch
```

## test

Run all tests.

```bash
go test -v ./...
```

## test-coverage

Run tests with coverage report.

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -html=coverage.out -o coverage.html
```

## run

Build and run the application.

```bash
go build -o sysdwitch -v ./cmd/sysdwitch
./service-control
```

## deps

Download and tidy dependencies.

```bash
go mod download
go mod tidy
```

## clean

Clean build artifacts.

```bash
go clean
rm -f service-control
rm -f service-control_unix
rm -f service-control-*
rm -f coverage.out coverage.html
```

## lint

Run linters (requires golangci-lint).

```bash
golangci-lint run
```

## fmt

Format Go code.

```bash
go fmt ./...
```

## vet

Run go vet.

```bash
go vet ./...
```

## check

Run format, vet, and lint checks.

```bash
go fmt ./...
go vet ./...
golangci-lint run
```
