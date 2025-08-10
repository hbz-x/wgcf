# OpenAPI Generated Client Documentation

This directory contains auto-generated Go client code for the Cloudflare WARP API, built from OpenAPI specifications using the OpenAPI Generator. This client provides programmatic access to Cloudflare's WARP service for WireGuard configuration and device management.

## Overview

The WGCF (WireGuard Cloudflare) project uses this OpenAPI-generated client to interact with Cloudflare's private API endpoints that power the Cloudflare WARP mobile application. The client handles device registration, account management, and WireGuard configuration retrieval.

### Generated Client Architecture

The client follows the standard OpenAPI Generator Go client structure:

```
openapi/
├── client.go              # Main API client and HTTP handling
├── configuration.go       # Client configuration and server settings
├── api_default.go         # DefaultApi service with all endpoints
├── model_*.go             # Request/response models
├── utils.go               # Utility functions and nullable types
├── response.go            # API response wrapper
└── docs/                  # Auto-generated documentation
```

## Core Components

### 1. APIClient (`client.go`)

The main `APIClient` struct manages HTTP communication with the Cloudflare API:

```go
type APIClient struct {
    cfg     *Configuration
    common  service
    DefaultApi *DefaultApiService
}
```

**Key Features:**
- HTTP request/response handling with proper error management
- Content-type detection (JSON/XML)
- Authentication support (OAuth2, Basic Auth, Bearer tokens)
- Debug mode for request/response logging
- Multipart form and file upload support
- Customizable HTTP client and timeouts

### 2. Configuration (`configuration.go`)

The `Configuration` struct manages client settings:

```go
type Configuration struct {
    Host             string
    Scheme           string
    DefaultHeader    map[string]string
    UserAgent        string
    Debug            bool
    Servers          ServerConfigurations
    OperationServers map[string]ServerConfigurations
    HTTPClient       *http.Client
}
```

**Configuration Options:**
- **Server Selection**: Multiple server configurations with templated URLs
- **Authentication Contexts**: Support for various auth methods via context
- **Custom Headers**: Default headers for all requests
- **HTTP Client**: Customizable transport and timeout settings

### 3. DefaultApi Service (`api_default.go`)

Contains all API endpoints as methods on the `DefaultApiService`:

| Method | Endpoint | Description |
|--------|----------|-------------|
| `Register` | `POST /{apiVersion}/reg` | Register a new device |
| `GetSourceDevice` | `GET /{apiVersion}/reg/{sourceDeviceId}` | Get device details |
| `UpdateSourceDevice` | `PATCH /{apiVersion}/reg/{sourceDeviceId}` | Update device settings |
| `GetAccount` | `GET /{apiVersion}/reg/{sourceDeviceId}/account` | Get account information |
| `UpdateAccount` | `PUT /{apiVersion}/reg/{sourceDeviceId}/account` | Update account (add license) |
| `GetBoundDevices` | `GET /{apiVersion}/reg/{sourceDeviceId}/account/devices` | List bound devices |
| `UpdateBoundDevice` | `PATCH /{apiVersion}/reg/{sourceDeviceId}/account/reg/{boundDeviceId}` | Update bound device |
| `ResetAccountLicense` | `POST /{apiVersion}/reg/{sourceDeviceId}/account/license` | Reset account license |
| `GetClientConfig` | `GET /{apiVersion}/client_config` | Get client configuration |

## API Models and Data Structures

### Request Models

**RegisterRequest** - Device registration:
```go
type RegisterRequest struct {
    FcmToken  string `json:"fcm_token"`   // Firebase Cloud Messaging token
    InstallId string `json:"install_id"`  // Installation identifier
    Key       string `json:"key"`         // WireGuard public key
    Locale    string `json:"locale"`      // Device locale (e.g., "en_US")
    Model     string `json:"model"`       // Device model name
    Tos       string `json:"tos"`         // Terms of service timestamp
    Type      string `json:"type"`        // Device type (e.g., "Android")
}
```

**UpdateAccountRequest** - License key updates:
```go
type UpdateAccountRequest struct {
    License string `json:"license"`      // WARP+ license key
}
```

### Response Models

**GetSourceDevice200Response** - Complete device information:
```go
type GetSourceDevice200Response struct {
    Account         GetSourceDevice200ResponseAccount `json:"account"`
    Config          GetSourceDevice200ResponseConfig  `json:"config"`
    Created         string `json:"created"`
    Enabled         bool   `json:"enabled"`
    FcmToken        string `json:"fcm_token"`
    Id              string `json:"id"`
    InstallId       string `json:"install_id"`
    Key             string `json:"key"`
    Locale          string `json:"locale"`
    Model           string `json:"model"`
    Name            string `json:"name"`
    Place           float32 `json:"place"`
    Tos             string `json:"tos"`
    Type            string `json:"type"`
    Updated         string `json:"updated"`
    WaitlistEnabled bool   `json:"waitlist_enabled"`
    WarpEnabled     bool   `json:"warp_enabled"`
}
```

**WireGuard Configuration** - Embedded in device response:
```go
type GetSourceDevice200ResponseConfig struct {
    ClientId  string                                        `json:"client_id"`
    Interface GetSourceDevice200ResponseConfigInterface    `json:"interface"`
    Peers     []GetSourceDevice200ResponseConfigPeers      `json:"peers"`
    Services  GetSourceDevice200ResponseConfigServices     `json:"services"`
}
```

## Usage Examples

### Basic Client Setup

```go
package main

import (
    "context"
    "fmt"
    "github.com/ViRb3/wgcf/v2/openapi"
)

func main() {
    // Create configuration
    config := openapi.NewConfiguration()
    config.Servers = []openapi.ServerConfiguration{
        {URL: "https://api.cloudflareclient.com"},
    }
    config.DefaultHeader = map[string]string{
        "User-Agent":        "okhttp/3.12.1",
        "CF-Client-Version": "a-6.3-1922",
    }
    
    // Create client
    client := openapi.NewAPIClient(config)
}
```

### Device Registration

```go
func registerDevice(client *openapi.APIClient, publicKey string) (*openapi.Register200Response, error) {
    registerReq := openapi.RegisterRequest{
        FcmToken:  "",           // Usually empty for non-mobile clients
        InstallId: "",           // Usually empty for non-mobile clients
        Key:       publicKey,    // WireGuard public key
        Locale:    "en_US",
        Model:     "PC",         // Device model identifier
        Tos:       "2023-01-01T00:00:00.000Z", // Current timestamp
        Type:      "Android",    // Client type
    }
    
    result, response, err := client.DefaultApi.
        Register(context.Background(), "v0a1922").
        RegisterRequest(registerReq).
        Execute()
        
    if err != nil {
        return nil, fmt.Errorf("registration failed: %v", err)
    }
    
    return &result, nil
}
```

### Authenticated Requests

```go
func getDeviceInfo(client *openapi.APIClient, deviceId, accessToken string) (*openapi.GetSourceDevice200Response, error) {
    // Add Bearer token to headers
    client.GetConfig().DefaultHeader["Authorization"] = "Bearer " + accessToken
    
    result, response, err := client.DefaultApi.
        GetSourceDevice(context.Background(), "v0a1922", deviceId).
        Execute()
        
    if err != nil {
        return nil, fmt.Errorf("failed to get device info: %v", err)
    }
    
    return &result, nil
}
```

### License Key Management

```go
func updateLicense(client *openapi.APIClient, deviceId, accessToken, licenseKey string) error {
    client.GetConfig().DefaultHeader["Authorization"] = "Bearer " + accessToken
    
    updateReq := openapi.UpdateAccountRequest{
        License: licenseKey,
    }
    
    result, response, err := client.DefaultApi.
        UpdateAccount(context.Background(), deviceId, "v0a1922").
        UpdateAccountRequest(updateReq).
        Execute()
        
    if err != nil {
        return fmt.Errorf("license update failed: %v", err)
    }
    
    fmt.Printf("License updated successfully: %+v\n", result)
    return nil
}
```

## Integration with Main Application

The WGCF application wraps this OpenAPI client in `/cloudflare/api.go` with higher-level convenience functions:

```go
// Wrapped client creation with WARP-specific defaults
func MakeApiClient(authToken *string) *openapi.APIClient {
    httpClient := http.Client{Transport: DefaultTransport}
    apiClient := openapi.NewAPIClient(&openapi.Configuration{
        DefaultHeader: DefaultHeaders,  // WARP-specific headers
        UserAgent:     DefaultHeaders["User-Agent"],
        Debug:         false,
        Servers: []openapi.ServerConfiguration{
            {URL: ApiUrl},  // https://api.cloudflareclient.com
        },
        HTTPClient: &httpClient,
    })
    if authToken != nil {
        apiClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + *authToken
    }
    return apiClient
}

// Type aliases for better integration
type Device = openapi.UpdateSourceDevice200Response
type Account = openapi.GetAccount200Response
type BoundDevice = openapi.GetBoundDevices200Response
```

## Error Handling

The client provides structured error handling through the `GenericOpenAPIError` type:

```go
type GenericOpenAPIError struct {
    body  []byte        // Raw response body
    error string        // Error message
    model interface{}   // Parsed error model (if applicable)
}
```

**Error Handling Example:**
```go
result, response, err := client.DefaultApi.GetAccount(ctx, deviceId, apiVersion).Execute()
if err != nil {
    if apiErr, ok := err.(openapi.GenericOpenAPIError); ok {
        fmt.Printf("API Error: %s\n", apiErr.Error())
        fmt.Printf("Response Body: %s\n", string(apiErr.Body()))
        fmt.Printf("Status Code: %d\n", response.StatusCode)
    }
    return nil, err
}
```

## Utility Functions

The client includes utility functions in `utils.go` for working with nullable types:

- **Pointer Helpers**: `PtrString()`, `PtrBool()`, `PtrInt()`, etc.
- **Nullable Types**: `NullableString`, `NullableBool`, etc. with JSON marshal/unmarshal support

**Example Usage:**
```go
updateReq := openapi.UpdateBoundDeviceRequest{
    Name:   openapi.PtrString("My Device"),  // Convert string to *string
    Active: openapi.PtrBool(true),           // Convert bool to *bool
}
```

## Generation Process

The OpenAPI client is generated using the OpenAPI Generator CLI tool:

### Generation Script (`generate-api.sh`)

```bash
#!/bin/bash

# Generate OpenAPI spec from API exploration
api generate:oas --json
mv ".optic/generated/openapi.json" "openapi-spec.json"

# Format the spec
go run "spec_format/main.go"

# Clean existing client
rm -rf "openapi"

# Generate Go client
openapi-generator-cli generate -i "openapi-spec.json" -g go -o "openapi"
```

### Generator Configuration

The client is generated with the following settings:
- **Generator**: `go` (OpenAPI Generator Go client)
- **Package**: `openapi`
- **API Version**: `536` (from Cloudflare's API)
- **Go Client Features**: Context support, proper error handling, nullable types

### Regeneration Process

1. **API Discovery**: Use Optic or similar tools to capture API calls and generate OpenAPI spec
2. **Spec Formatting**: Clean and format the OpenAPI specification
3. **Code Generation**: Run OpenAPI Generator to create Go client
4. **Integration**: Update wrapper functions in `/cloudflare/api.go` if needed

## Customization and Extension

### Custom HTTP Transport

The main application customizes the HTTP transport for Cloudflare's requirements:

```go
DefaultTransport = &http.Transport{
    // Match app's TLS config or API will reject with 403 error 1020
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        MaxVersion: tls.VersionTLS12,
    },
    ForceAttemptHTTP2: false,
    // Standard transport settings...
}
```

### Authentication Management

Authentication is handled through:
- **Bearer Tokens**: Set in `Authorization` header for authenticated endpoints
- **Context Values**: Support for OAuth2 and other auth methods
- **Global Client**: Shared authenticated client instance in the wrapper

### Adding New Endpoints

To add support for new API endpoints:

1. **Update OpenAPI Spec**: Add new paths/operations to `openapi-spec.json`
2. **Regenerate Client**: Run `./generate-api.sh`
3. **Add Wrapper Functions**: Create convenience functions in `/cloudflare/api.go`
4. **Update Types**: Add type aliases if needed for better integration

## Best Practices

### Client Configuration

- **Timeouts**: Set appropriate HTTP timeouts for network conditions
- **Retry Logic**: Implement retry logic in wrapper functions for transient failures
- **Rate Limiting**: Respect API rate limits (not currently implemented)
- **Error Logging**: Enable debug mode during development for request/response logging

### Security Considerations

- **Token Storage**: Never log or expose access tokens
- **TLS Configuration**: Use proper TLS settings as required by Cloudflare
- **Input Validation**: Validate inputs before making API calls
- **Error Messages**: Don't expose sensitive information in error messages

### Performance Optimization

- **Connection Reuse**: Use persistent HTTP connections
- **Client Reuse**: Reuse APIClient instances rather than creating new ones
- **Minimal Data**: Only request necessary fields when possible
- **Batch Operations**: Use batch endpoints when available

## Troubleshooting

### Common Issues

1. **403 Forbidden (Error 1020)**
   - Cause: Incorrect TLS configuration or missing headers
   - Solution: Ensure TLS 1.2 and proper User-Agent header

2. **Authentication Failures**
   - Cause: Missing or invalid Bearer token
   - Solution: Verify token format and expiration

3. **JSON Parsing Errors**
   - Cause: API response format changes
   - Solution: Regenerate client from updated OpenAPI spec

4. **Network Timeouts**
   - Cause: Slow network or restrictive firewall
   - Solution: Increase timeout values in HTTP client

### Debug Mode

Enable debug mode to see raw HTTP requests/responses:

```go
config := openapi.NewConfiguration()
config.Debug = true
client := openapi.NewAPIClient(config)
```

## API Reference Links

- **OpenAPI Specification**: `/openapi/api/openapi.yaml`
- **Generated Documentation**: `/openapi/docs/`
- **API Endpoints**: `/openapi/docs/DefaultApi.md`
- **Model Documentation**: `/openapi/docs/*.md`

## Conclusion

This OpenAPI-generated client provides a robust, type-safe interface to Cloudflare's WARP API. The generated code handles the low-level HTTP details while the wrapper layer provides convenient, application-specific functions. The combination enables reliable programmatic access to WARP functionality for WireGuard configuration management.

The client follows Go best practices for HTTP clients and integrates seamlessly with the broader WGCF application architecture. Regular regeneration ensures compatibility with API changes while maintaining a stable interface for the application layer.