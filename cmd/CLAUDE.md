# WGCF CLI Command Structure

This document provides a comprehensive guide to the CLI command structure for the `wgcf` (WireGuard Cloudflare) utility, built using the Cobra framework.

## Project Overview

WGCF is a utility for Cloudflare Warp that allows users to create and manage accounts, assign license keys, and generate WireGuard profiles. The CLI is organized using the Cobra framework with a clear hierarchical command structure.

## Command Structure and Organization

### Root Command (`/root/wgcf/cmd/root.go`)

The root command serves as the entry point and coordinator for all subcommands:

```go
var RootCmd = &cobra.Command{
    Use:   "wgcf",
    Short: "WireGuard Cloudflare Warp utility",
    Long: FormatMessage("", `...`),
    Run: func(cmd *cobra.Command, args []string) {
        // Shows help when run without subcommands
    },
}
```

**Key Features:**
- **Configuration Management**: Uses Viper for configuration with default file `wgcf-account.toml`
- **Environment Variables**: Supports env vars with `WGCF_` prefix
- **Global Flags**: `--config` flag for specifying custom configuration file
- **Auto-initialization**: Calls `initConfig()` on startup

### Available Commands

| Command | Purpose | File Location |
|---------|---------|---------------|
| `register` | Register new Cloudflare Warp device | `/root/wgcf/cmd/register/register.go` |
| `update` | Update existing account | `/root/wgcf/cmd/update/update.go` |
| `generate` | Generate WireGuard profile | `/root/wgcf/cmd/generate/generate.go` |
| `status` | Show device status | `/root/wgcf/cmd/status/status.go` |
| `trace` | Show connection trace info | `/root/wgcf/cmd/trace/trace.go` |

## Command Execution Flow

### 1. Application Startup (`main.go`)
```go
func main() {
    if err := cmd.Execute(); err != nil {
        log.Fatal(util.GetErrorMessage(err))
    }
}
```

### 2. Root Command Initialization
1. **Configuration Setup**: `initConfig()` loads configuration from file/environment
2. **Command Registration**: All subcommands are added to root in `init()`
3. **Flag Processing**: Global and command-specific flags are processed

### 3. Command-Specific Execution
Each command follows this pattern:
1. **Validation**: Check prerequisites (e.g., account exists for certain commands)
2. **API Interaction**: Use Cloudflare API via the `/cloudflare` package  
3. **Configuration Updates**: Update local config file using Viper
4. **Output**: Display results and status information

## Configuration Handling

### Configuration Structure (`/root/wgcf/config/config.go`)

```go
const (
    DeviceId    = "device_id"
    AccessToken = "access_token"  
    PrivateKey  = "private_key"
    LicenseKey  = "license_key"
)

type Context struct {
    DeviceId    string
    AccessToken string
    PrivateKey  string
    LicenseKey  string
}
```

### Configuration Flow

1. **Defaults**: Set via `initConfigDefaults()` in root.go
2. **File Loading**: TOML config file loaded by Viper
3. **Environment Override**: `WGCF_*` environment variables
4. **Context Creation**: `CreateContext()` creates API context from config

### Configuration Access Pattern

```go
// Reading configuration
deviceId := viper.GetString(config.DeviceId)

// Writing configuration
viper.Set(config.DeviceId, newDeviceId)
viper.WriteConfig()
```

## Shared Functionality and Patterns

### Shared Utilities (`/root/wgcf/cmd/shared/shared.go`)

#### Core Functions

| Function | Purpose |
|----------|---------|
| `FormatMessage()` | Standardized message formatting for command descriptions |
| `IsConfigValidAccount()` | Validates if account configuration is complete |
| `CreateContext()` | Creates API context from Viper configuration |
| `F32ToHumanReadable()` | Converts bytes to human-readable format (KB, MB, etc.) |
| `PrintDeviceData()` | Standardized device information display |
| `SetDeviceName()` | Updates device name via API |

#### Validation Pattern
```go
if !IsConfigValidAccount() {
    return errors.New("no account detected")
}
```

#### Context Creation Pattern
```go
ctx := CreateContext()
thisDevice, err := cloudflare.GetSourceDevice(ctx)
```

### Common Command Patterns

#### 1. Command Structure Template
```go
var shortMsg = "Command description"

var Cmd = &cobra.Command{
    Use:   "commandname",
    Short: shortMsg,
    Long:  FormatMessage(shortMsg, `extended description`),
    Run: func(cmd *cobra.Command, args []string) {
        if err := commandFunction(); err != nil {
            log.Fatal(util.GetErrorMessage(err))
        }
    },
}

func init() {
    // Add command-specific flags
    Cmd.PersistentFlags().StringVarP(&variable, "flag", "f", "default", "description")
}
```

#### 2. Error Handling Pattern
All commands use consistent error handling:
```go
if err := someOperation(); err != nil {
    return err  // or log.Fatal(util.GetErrorMessage(err))
}
```

#### 3. API Interaction Pattern
```go
ctx := CreateContext()
result, err := cloudflare.SomeAPICall(ctx)
if err != nil {
    return err
}
// Process result
```

## Adding New Commands

### Step-by-Step Guide

1. **Create Command Directory**
   ```bash
   mkdir /root/wgcf/cmd/newcommand
   ```

2. **Create Command File**: `/root/wgcf/cmd/newcommand/newcommand.go`
   ```go
   package newcommand

   import (
       . "github.com/ViRb3/wgcf/v2/cmd/shared"
       "github.com/ViRb3/wgcf/v2/util"
       "github.com/spf13/cobra"
       "log"
   )

   var shortMsg = "Your command description"

   var Cmd = &cobra.Command{
       Use:   "newcommand",
       Short: shortMsg,
       Long:  FormatMessage(shortMsg, `Extended description here`),
       Run: func(cmd *cobra.Command, args []string) {
           if err := executeNewCommand(); err != nil {
               log.Fatal(util.GetErrorMessage(err))
           }
       },
   }

   func init() {
       // Add flags if needed
   }

   func executeNewCommand() error {
       // Command logic here
       return nil
   }
   ```

3. **Register Command in Root** (`/root/wgcf/cmd/root.go`):
   ```go
   import (
       "github.com/ViRb3/wgcf/v2/cmd/newcommand"
   )
   
   func init() {
       // ... existing code ...
       RootCmd.AddCommand(newcommand.Cmd)
   }
   ```

### Command Implementation Guidelines

#### Required Imports
- Always import shared utilities: `. "github.com/ViRb3/wgcf/v2/cmd/shared"`
- Import util for error handling: `"github.com/ViRb3/wgcf/v2/util"`
- Import Cobra: `"github.com/spf13/cobra"`

#### Validation Requirements
- Use `IsConfigValidAccount()` if command requires existing account
- Validate flags and parameters early in execution
- Return descriptive errors for validation failures

#### API Integration
- Create context using `CreateContext()`
- Use existing Cloudflare API functions from `/cloudflare` package
- Handle API errors gracefully

#### Output Standards
- Use `log.Println()` for informational messages
- Use `PrintDeviceData()` for device information display
- Use `log.Fatal(util.GetErrorMessage(err))` for fatal errors

## Key Functions and Their Purposes

### Root Command Functions

| Function | Purpose |
|----------|---------|
| `Execute()` | Entry point called from main() |
| `initConfig()` | Loads configuration from file and environment |
| `initConfigDefaults()` | Sets default configuration values |

### Shared Utility Functions

| Function | Purpose |
|----------|---------|
| `FormatMessage()` | Formats short and long command descriptions |
| `IsConfigValidAccount()` | Validates account configuration completeness |
| `CreateContext()` | Creates API context from current configuration |
| `PrintDeviceData()` | Displays standardized device and account information |
| `SetDeviceName()` | Updates device name through API |
| `F32ToHumanReadable()` | Converts numeric values to human-readable format |

### Command-Specific Functions

#### Register Command
- `registerAccount()`: Complete account registration flow
- `checkTOS()`: Terms of Service acceptance handling

#### Update Command  
- `updateAccount()`: Updates existing account
- `ensureLicenseKeyUpToDate()`: Checks and updates license key
- `updateLicenseKey()`: Handles license key changes

#### Generate Command
- `generateProfile()`: Creates WireGuard configuration file

#### Status Command
- `status()`: Displays current device and account status

#### Trace Command
- `trace()`: Shows connection trace information

## Architecture Design Principles

### 1. Separation of Concerns
- **Commands**: Handle CLI interface and user interaction
- **Shared**: Common utilities and validation
- **Cloudflare**: API interactions
- **Config**: Configuration management
- **Util**: Generic utility functions

### 2. Consistent Error Handling
- All commands use `util.GetErrorMessage()` for error formatting
- Consistent error propagation pattern
- Fatal errors handled at command level

### 3. Configuration Management
- Centralized through Viper
- Environment variable support
- File-based persistence
- Default value handling

### 4. Modular Command Structure
- Each command in separate package
- Standardized command interface
- Shared functionality through imports

### 5. API Context Pattern
- Consistent context creation from configuration
- Centralized API credential management
- Reusable across all commands

This architecture provides a solid foundation for extending the CLI with new commands while maintaining consistency and reliability across the codebase.