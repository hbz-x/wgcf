# Cloudflare API Integration

This document provides a comprehensive analysis of the Cloudflare API integration in the `cloudflare/` directory, which handles communication with Cloudflare Warp services for WireGuard configuration.

## Overview

The cloudflare package serves as a high-level wrapper around the OpenAPI-generated client, providing specialized functionality for interacting with Cloudflare's Warp API. It handles device registration, account management, and WireGuard configuration retrieval.

## Architecture

### Directory Structure

```
cloudflare/
├── api.go    - Main API client and high-level methods
└── util.go   - Utility functions for device management
```

### Dependencies

The package integrates with several internal modules:
- `openapi/` - Auto-generated API client from OpenAPI specification
- `config/` - Configuration context management
- `util/` - Common utilities (timestamps, restructuring)
- `wireguard/` - WireGuard key management

## API Client Setup (`api.go`)

### Base Configuration

```go
const (
    ApiUrl     = "https://api.cloudflareclient.com"
    ApiVersion = "v0a1922"
)
```

The API version `v0a1922` appears to correspond to a specific Cloudflare client version, ensuring compatibility with their backend services.

### Default Headers

```go
var DefaultHeaders = map[string]string{
    "User-Agent":        "okhttp/3.12.1",
    "CF-Client-Version": "a-6.3-1922",
}
```

**Critical Implementation Detail**: The client mimics the official Cloudflare mobile app by using:
- **User-Agent**: `okhttp/3.12.1` (Android HTTP client)
- **CF-Client-Version**: `a-6.3-1922` (Android client version 6.3 build 1922)

This impersonation is necessary because Cloudflare's API validates client authenticity and will reject requests from unrecognized clients.

## HTTP Transport and TLS Configuration

### Custom TLS Settings

```go
var DefaultTransport = &http.Transport{
    // Match app's TLS config or API will reject us with code 403 error 1020
    TLSClientConfig: &tls.Config{
        MinVersion: tls.VersionTLS12,
        MaxVersion: tls.VersionTLS12
    },
    ForceAttemptHTTP2: false,
    // Standard HTTP transport settings...
}
```

### Why TLS 1.2 Only?

The TLS configuration is **critically important** and locked to TLS 1.2 for the following reasons:

1. **API Fingerprinting**: Cloudflare uses TLS fingerprinting as part of their bot detection
2. **Error 1020 Prevention**: Without exact TLS version matching, the API returns HTTP 403 with Cloudflare error 1020
3. **Client Validation**: The official mobile app uses TLS 1.2, so the wrapper must match exactly
4. **HTTP/2 Disabled**: `ForceAttemptHTTP2: false` prevents protocol negotiation that might expose the client as non-mobile

### Transport Settings

Standard HTTP transport settings are preserved from `http.DefaultTransport`:
- **Proxy Support**: `http.ProxyFromEnvironment`
- **Connection Pooling**: 100 max idle connections
- **Timeouts**: 90s idle, 10s TLS handshake, 1s expect-continue

## Authentication Mechanisms

### Bearer Token Authentication

```go
if authToken != nil {
    apiClient.GetConfig().DefaultHeader["Authorization"] = "Bearer " + *authToken
}
```

Authentication uses Bearer tokens in the Authorization header. The system supports two types of clients:

1. **Anonymous Client**: Used for device registration (no auth token)
2. **Authenticated Client**: Used for device management (with access token)

### Global Client Management

```go
var apiClient = MakeApiClient(nil)           // Anonymous client
var apiClientAuth *openapi.APIClient        // Cached authenticated client

func globalClientAuth(authToken string) *openapi.APIClient {
    if apiClientAuth == nil {
        apiClientAuth = MakeApiClient(&authToken)
    }
    return apiClientAuth
}
```

**Singleton Pattern**: The system maintains a global authenticated client instance to avoid recreating HTTP connections and configurations.

## Available API Methods

### Device Registration

```go
func Register(publicKey *wireguard.Key, deviceModel string) (openapi.Register200Response, error)
```

**Purpose**: Registers a new device with Cloudflare Warp service
**Authentication**: None required (anonymous)
**Parameters**:
- `publicKey`: WireGuard public key for the device
- `deviceModel`: Device identifier (e.g., "PC", mobile device name)

**Request Structure**:
```go
RegisterRequest{
    FcmToken:  "",           // Firebase token (empty for non-mobile)
    InstallId: "",           // Installation ID (empty for non-mobile) 
    Key:       publicKey.String(),
    Locale:    "en_US",
    Model:     deviceModel,
    Tos:       timestamp,    // Terms of Service acceptance timestamp
    Type:      "Android",    // Always "Android" to match mobile client
}
```

### Device Information Retrieval

```go
func GetSourceDevice(ctx *config.Context) (*Device, error)
```

**Purpose**: Retrieves complete device configuration including WireGuard settings
**Authentication**: Required (uses access token)
**Returns**: Device configuration with account details and WireGuard config

### Account Management

```go
func GetAccount(ctx *config.Context) (*Account, error)
func UpdateLicenseKey(ctx *config.Context) (*openapi.UpdateAccount200Response, error)
```

**Purpose**: Account information retrieval and license key management
**Authentication**: Required
**Key Features**:
- Account type and premium status
- Data quota and usage information
- License key updates for premium features

### Device Binding and Management

```go
func GetBoundDevices(ctx *config.Context) ([]BoundDevice, error)
func GetSourceBoundDevice(ctx *config.Context) (*BoundDevice, error)
func UpdateSourceBoundDeviceName(ctx *config.Context, newName string) (*BoundDevice, error)
func UpdateSourceBoundDeviceActive(ctx *config.Context, active bool) (*BoundDevice, error)
```

**Purpose**: Manage devices associated with an account
**Authentication**: Required
**Capabilities**:
- List all bound devices
- Find current device in bound device list
- Update device name and active status
- Device activation/deactivation

## Data Models and Type Casting

### Type Aliases

```go
type Device openapi.UpdateSourceDevice200Response
type Account openapi.GetAccount200Response  
type BoundDevice openapi.GetBoundDevices200Response
```

**Design Pattern**: The package creates type aliases for OpenAPI-generated structs to provide cleaner interfaces and potential future customization points.

### Data Restructuring

```go
func GetSourceDevice(ctx *config.Context) (*Device, error) {
    result, _, err := globalClientAuth(ctx.AccessToken).DefaultApi.
        GetSourceDevice(nil, ApiVersion, ctx.DeviceId).Execute()
    
    castResult := Device{}
    if err := util.Restructure(&result, &castResult); err != nil {
        return nil, err
    }
    return &castResult, err
}
```

**Restructuring Pattern**: Uses YAML marshal/unmarshal for reliable type conversion between OpenAPI-generated types and custom aliases.

## Error Handling Patterns

### HTTP Error Propagation

The package follows Go's standard error handling patterns:

1. **Error Forwarding**: HTTP and parsing errors are forwarded directly from the OpenAPI client
2. **Context Validation**: Methods assume valid authentication context
3. **Business Logic Errors**: Custom errors for business logic (e.g., device not found)

### Example Error Cases

- **HTTP 403 + Error 1020**: TLS fingerprint mismatch or invalid client headers
- **HTTP 401**: Invalid or expired access token
- **HTTP 404**: Device or account not found
- **Device Not Found**: Custom error when device ID doesn't match any bound device

## Utility Functions (`util.go`)

### Device Search

```go
func FindDevice(devices []BoundDevice, deviceId string) (*BoundDevice, error) {
    for _, device := range devices {
        if device.Id == deviceId {
            return &device, nil
        }
    }
    return nil, errors.New("device not found in list")
}
```

**Purpose**: Linear search through bound devices by ID
**Error Handling**: Returns structured error when device not found

## Integration with OpenAPI Generated Client

### Client Configuration

The package leverages the OpenAPI-generated client with custom configuration:

```go
apiClient := openapi.NewAPIClient(&openapi.Configuration{
    DefaultHeader: DefaultHeaders,
    UserAgent:     DefaultHeaders["User-Agent"],
    Debug:         false,
    Servers: []openapi.ServerConfiguration{
        {URL: ApiUrl},
    },
    HTTPClient: &httpClient,
})
```

### API Method Mapping

Each high-level method corresponds to OpenAPI-generated methods:

- `Register()` → `DefaultApi.Register()`
- `GetSourceDevice()` → `DefaultApi.GetSourceDevice()`
- `GetAccount()` → `DefaultApi.GetAccount()`
- `UpdateAccount()` → `DefaultApi.UpdateAccount()`
- `GetBoundDevices()` → `DefaultApi.GetBoundDevices()`
- `UpdateBoundDevice()` → `DefaultApi.UpdateBoundDevice()`

## How to Add New API Endpoints

To extend the API client with new endpoints:

### 1. Update OpenAPI Specification

Add new endpoints to `/root/wgcf/openapi-spec.json`:

```json
{
  "paths": {
    "/v0a1922/reg/{sourceDeviceId}/new-endpoint": {
      "get": {
        "summary": "New endpoint description",
        "parameters": [...],
        "responses": {...}
      }
    }
  }
}
```

### 2. Regenerate OpenAPI Client

```bash
./generate-api.sh
```

This will update the `openapi/` directory with new generated code.

### 3. Add High-Level Wrapper

In `cloudflare/api.go`, add a wrapper method:

```go
func NewEndpoint(ctx *config.Context, param string) (*NewResponse, error) {
    result, _, err := globalClientAuth(ctx.AccessToken).DefaultApi.
        NewEndpoint(nil, ApiVersion, ctx.DeviceId, param).
        Execute()
    if err != nil {
        return nil, err
    }
    castResult := NewResponse(result)
    return &castResult, nil
}
```

### 4. Create Type Alias

Add type alias for the response:

```go
type NewResponse openapi.GetNewEndpoint200Response
```

### 5. Add Utility Functions (if needed)

In `cloudflare/util.go`, add any helper functions for data manipulation.

## Security Considerations

### API Key Management

- Access tokens are managed through the `config.Context` structure
- Tokens are stored in application configuration, not hardcoded
- No token validation or refresh logic (assumes external management)

### TLS Security

- Forced TLS 1.2 provides strong encryption
- Certificate validation enabled by default
- No custom certificate handling or bypassing

### Client Impersonation

The client impersonates the official Cloudflare mobile app, which raises considerations:
- **Terms of Service**: Ensure usage complies with Cloudflare's ToS
- **Rate Limiting**: Respect API rate limits to avoid blocking
- **API Stability**: Cloudflare may change client validation at any time

## Usage Examples

### Device Registration Flow

```go
// Generate WireGuard key pair
privateKey, err := wireguard.NewPrivateKey()
if err != nil {
    return err
}

// Register device
device, err := cloudflare.Register(privateKey.Public(), "PC")
if err != nil {
    return err
}

// Store configuration
viper.Set(config.DeviceId, device.Id)
viper.Set(config.AccessToken, device.Token)
viper.Set(config.PrivateKey, privateKey.String())
```

### Device Status Check

```go
ctx := &config.Context{
    DeviceId:    viper.GetString(config.DeviceId),
    AccessToken: viper.GetString(config.AccessToken),
}

device, err := cloudflare.GetSourceDevice(ctx)
if err != nil {
    return err
}

boundDevice, err := cloudflare.GetSourceBoundDevice(ctx)
if err != nil {
    return err
}

fmt.Printf("Device: %s, Active: %v\n", device.Name, boundDevice.Active)
```

### Account Information

```go
account, err := cloudflare.GetAccount(ctx)
if err != nil {
    return err
}

fmt.Printf("Account Type: %s, Warp+: %v\n", 
    account.AccountType, account.WarpPlus)
```

## Best Practices

### 1. Context Management

Always use proper `config.Context` with valid tokens:

```go
func CreateContext() *config.Context {
    return &config.Context{
        DeviceId:    viper.GetString(config.DeviceId),
        AccessToken: viper.GetString(config.AccessToken),
        PrivateKey:  viper.GetString(config.PrivateKey),
        LicenseKey:  viper.GetString(config.LicenseKey),
    }
}
```

### 2. Error Handling

Implement comprehensive error handling for API failures:

```go
device, err := cloudflare.GetSourceDevice(ctx)
if err != nil {
    // Log the full error context
    log.Printf("Failed to get device: %+v", err)
    return err
}
```

### 3. Client Reuse

Use the global client instances to avoid connection overhead:

```go
// Good - reuses connection
client := globalClientAuth(ctx.AccessToken)

// Avoid - creates new connections
client := MakeApiClient(&ctx.AccessToken)
```

### 4. Rate Limiting

Implement appropriate delays between API calls:

```go
time.Sleep(time.Millisecond * 100) // Small delay between requests
```

## Future Improvements

### Potential Enhancements

1. **Token Refresh**: Automatic access token refresh on expiration
2. **Retry Logic**: Exponential backoff for transient failures
3. **Caching**: Response caching for frequently accessed data
4. **Metrics**: API call timing and success rate tracking
5. **Health Checks**: Periodic connectivity and authentication validation

### API Evolution

Monitor Cloudflare's client updates for changes in:
- Client version headers
- TLS requirements
- Authentication mechanisms
- New API endpoints

The current implementation is tied to specific client version `a-6.3-1922` and may require updates as Cloudflare evolves their mobile applications.