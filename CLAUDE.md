# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Testing
- `go build .` - Build the main binary
- `go test ./...` - Run all tests including unit tests
- `docker buildx build --load --tag wgcf:test --file ./Dockerfile ./` - Build Docker image for testing

### API Documentation Generation
- `api start` - Start Optic API documentation capture (requires [Optic](https://github.com/opticdev/optic) installed)
- `bash generate-api.sh` - Regenerate Go client API code from OpenAPI spec (requires [openapi-generator](https://openapi-generator.tech/) installed)

The API generation process:
1. Run tests in `api_tests/main.go` with Optic to capture API calls
2. Generate OpenAPI3 specification at `openapi-spec.json`
3. Format spec with `spec_format/main.go`  
4. Generate Go client code in `openapi/` directory

## Architecture

### Core Components
- **Main Entry Point**: `main.go` - Simple CLI entry delegating to cobra commands
- **CLI Framework**: Uses [Cobra](https://github.com/spf13/cobra) with commands in `cmd/` directory
- **Configuration**: TOML-based config via [Viper](https://github.com/spf13/viper), stored in `wgcf-account.toml`
- **API Client**: Auto-generated from OpenAPI spec in `openapi/` directory
- **WireGuard**: Profile generation and key management in `wireguard/`

### Command Structure
All CLI commands are in `cmd/` with consistent structure:
- `cmd/root.go` - Root command and config initialization
- `cmd/{command}/{command}.go` - Individual commands (register, generate, status, trace, update)
- `cmd/shared/shared.go` - Shared command utilities

### Key Modules
- **`cloudflare/`** - Cloudflare Warp API integration with custom HTTP transport for TLS compatibility
- **`wireguard/`** - WireGuard profile templating and key generation using `golang.org/x/crypto`
- **`config/`** - Configuration constants and context structure
- **`util/`** - Utility functions for error handling and common operations

### Authentication Flow
1. `register` - Creates new Cloudflare Warp device and account
2. `update` - Updates device with license key for Warp+ (optional)
3. `generate` - Creates WireGuard profile from device configuration
4. `status`/`trace` - Check account and connection status

### API Integration
- Uses auto-generated OpenAPI client targeting `https://api.cloudflareclient.com`
- Custom HTTP transport with TLS 1.2 constraint to match official app behavior
- Mimics official Android app headers (`CF-Client-Version: a-6.3-1922`, `User-Agent: okhttp/3.12.1`)

### Configuration Management
- Device credentials stored in TOML format: `device_id`, `access_token`, `private_key`, `license_key`
- Environment variable support with `WGCF_` prefix
- Default config file: `wgcf-account.toml`

### Testing
- Unit tests for utilities (`util/util_test.go`) and WireGuard components (`wireguard/*_test.go`)
- API integration tests in `api_tests/main.go` used for documentation generation
- CI runs tests and Docker build verification