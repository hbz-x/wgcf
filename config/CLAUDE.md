# Config Package Documentation

The `config` package provides the foundational configuration structure for the wgcf application, defining constants and context types for managing Cloudflare WARP device credentials.

## Configuration Constants

The package defines four key configuration constants that serve as keys for the Viper configuration system:

```go
const (
    DeviceId    = "device_id"    // Unique identifier for the registered device
    AccessToken = "access_token" // Authentication token for API access  
    PrivateKey  = "private_key"  // WireGuard private key for the device
    LicenseKey  = "license_key"  // Cloudflare WARP+ license key (optional)
)
```

### Configuration Constant Usage

- **DeviceId**: Used to identify the device when making API calls to Cloudflare
- **AccessToken**: Required for authenticating API requests
- **PrivateKey**: The WireGuard private key used for tunnel configuration
- **LicenseKey**: Optional premium license key for WARP+ features

## Context Structure

The `Context` struct provides a typed representation of the configuration values:

```go
type Context struct {
    DeviceId    string
    AccessToken string
    PrivateKey  string
    LicenseKey  string
}
```

This context object is used throughout the application to pass configuration data to API functions and other components that need access to device credentials.

## Viper Integration

The configuration integrates seamlessly with the Viper configuration management library:

### Configuration File Setup

- **Default file**: `wgcf-account.toml`
- **Format**: TOML (Tom's Obvious, Minimal Language)
- **Environment prefix**: `WGCF_` (allows environment variable overrides)

### Initialization Pattern

The configuration is initialized in `/root/wgcf/cmd/root.go`:

```go
func initConfigDefaults() {
    viper.SetDefault(config.DeviceId, "")
    viper.SetDefault(config.AccessToken, "")
    viper.SetDefault(config.PrivateKey, "")
    viper.SetDefault(config.LicenseKey, "")
}
```

### Reading Configuration Values

Configuration values are accessed using Viper's getter functions with the config constants:

```go
// Example from shared.go
func CreateContext() *config.Context {
    ctx := config.Context{
        DeviceId:    viper.GetString(config.DeviceId),
        AccessToken: viper.GetString(config.AccessToken),
        LicenseKey:  viper.GetString(config.LicenseKey),
    }
    return &ctx
}
```

### Validation Pattern

The application includes validation to ensure required configuration is present:

```go
func IsConfigValidAccount() bool {
    return viper.GetString(config.DeviceId) != "" &&
           viper.GetString(config.AccessToken) != "" &&
           viper.GetString(config.PrivateKey) != ""
}
```

### Setting Configuration Values

Configuration values are typically set during device registration:

```go
// Example from register.go
viper.Set(config.PrivateKey, privateKey.String())
viper.Set(config.DeviceId, device.Id)
viper.Set(config.AccessToken, device.Token)
viper.Set(config.LicenseKey, device.Account.License)
```

## Usage Patterns

### Creating a Context

The most common pattern is to create a context object for API calls:

```go
ctx := shared.CreateContext()
device, err := cloudflare.GetSourceDevice(ctx)
```

### Validation Before Operations

Always validate configuration before performing operations that require authentication:

```go
if !shared.IsConfigValidAccount() {
    return errors.New("no account detected")
}
```

### Environment Variable Override

Configuration can be overridden using environment variables with the `WGCF_` prefix:

- `WGCF_DEVICE_ID`
- `WGCF_ACCESS_TOKEN`
- `WGCF_PRIVATE_KEY`
- `WGCF_LICENSE_KEY`

## Integration Points

The config package is used by:

- **API Layer** (`/root/wgcf/cloudflare/`): All API functions accept a `*config.Context`
- **Command Layer** (`/root/wgcf/cmd/`): Commands use config validation and context creation
- **WireGuard Integration**: Private key from config is used for profile generation

## Best Practices

1. **Always validate configuration** before performing operations that require authentication
2. **Use the Context struct** rather than accessing Viper directly in business logic
3. **Handle missing configuration gracefully** with meaningful error messages
4. **Use environment variables** for CI/CD and automated deployments
5. **Keep sensitive data secure** - the configuration file contains credentials

## File Structure

```
config/
├── config.go          # Configuration constants and Context struct
└── CLAUDE.md          # This documentation
```

The config package serves as the foundation for credential management in wgcf, providing a clean interface between the Viper configuration system and the rest of the application.