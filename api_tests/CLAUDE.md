# API Tests for Documentation Generation

This directory contains API testing infrastructure specifically designed for generating OpenAPI documentation using the Optic platform. The tests simulate real API interactions to capture request/response patterns and automatically generate accurate API specifications.

## Overview

The `api_tests/` directory implements an automated API documentation generation workflow that:

1. **Captures Live API Calls**: Uses Optic to intercept and record HTTP requests/responses during test execution
2. **Generates OpenAPI Specs**: Automatically creates OpenAPI 3.0 specifications from captured API interactions
3. **Validates API Behavior**: Ensures APIs work correctly while building documentation
4. **Maintains Documentation Currency**: Keeps API docs synchronized with actual implementation

## Architecture

### Core Components

**`main.go`** - Single comprehensive test file that:
- Sets up Optic testing infrastructure with custom HTTP transport
- Implements complete API workflow testing (registration, authentication, device management)
- Captures all API endpoint interactions for specification generation
- Handles error scenarios and response validation

### Key Dependencies

- **`github.com/ViRb3/optic-go`** - Go SDK for Optic API documentation platform
- **`github.com/ViRb3/sling/v2`** - HTTP client for making API requests during tests
- **Custom Transport Layer** - Mimics production behavior with proper TLS and headers

## Test Structure and Flow

### 1. Optic Configuration

```go
testConfig := opticgo.Config{
    ApiUrl:          cloudflare.ApiUrl + cloudflare.ApiVersion, // Target API
    OpticUrl:        "http://localhost:8889",                   // Optic proxy server
    ProxyListenAddr: "localhost",                               // Proxy binding
    DebugPrint:      true,                                      // Verbose logging
    TripFunc:        CustomTransport{},                         // HTTP transport wrapper
    InternetCheckTimeout: 10 * time.Second,                     // Network validation
}
```

**Configuration Details:**
- **ApiUrl**: Points to Cloudflare WARP API (`https://api.cloudflareclient.com/v0a1922`)
- **OpticUrl**: Local Optic proxy that captures API interactions
- **ProxyListenAddr**: Network interface for request interception
- **TripFunc**: Custom HTTP transport that adds authentication and validation

### 2. Custom HTTP Transport

The `CustomTransport` struct provides critical functionality:

```go
type CustomTransport struct{}

func (CustomTransport) RoundTrip(r *http.Request) (*http.Response, error) {
    // Add production headers to mimic real client behavior
    for key, val := range defaultHeaders {
        r.Header.Set(key, val)
    }
    
    // Execute request through default transport
    response, err := defaultTransport.RoundTrip(r)
    if err != nil {
        return nil, err
    }
    
    // Validate successful responses (200 OK only)
    if response.StatusCode != 200 {
        return nil, errors.New(fmt.Sprintf("bad code: %d", response.StatusCode))
    }
    
    return response, nil
}
```

**Transport Features:**
- **Header Injection**: Automatically adds Cloudflare-specific headers (`CF-Client-Version`, `User-Agent`)
- **Authentication**: Injects Bearer tokens for authenticated endpoints
- **Response Validation**: Ensures only successful responses (200 OK) are processed
- **Error Handling**: Converts HTTP errors to Go errors for test framework

### 3. Test Definitions

The test suite implements a comprehensive API workflow:

```go
var tests = []opticgo.TestDefinition{
    {
        "get device",                                    // Human-readable test name
        nil,                                            // No request body
        fmt.Sprintf("reg/%s", deviceId),                // Endpoint path
        "GET",                                          // HTTP method
    },
    {
        "set device active",
        struct { Active bool `json:"active"` }{true},   // Request payload
        fmt.Sprintf("reg/%s/account/reg/%s", deviceId, deviceId),
        "PATCH",
    },
    // ... additional test definitions
}
```

**Test Workflow Coverage:**

1. **Device Registration**: `POST /reg` - Create new device account
2. **Device Retrieval**: `GET /reg/{deviceId}` - Fetch device details
3. **Account Management**: `GET /reg/{deviceId}/account` - Account information
4. **Device Listing**: `GET /reg/{deviceId}/account/devices` - Bound devices
5. **Device Updates**: `PATCH /reg/{deviceId}/account/reg/{deviceId}` - Modify device settings
6. **Client Configuration**: `GET /client_config` - Global client settings
7. **License Management**: `PUT /reg/{deviceId}/account` - License key updates
8. **Key Rotation**: `PATCH /reg/{deviceId}` - Update WireGuard keys
9. **License Reset**: `POST /reg/{deviceId}/account/license` - Generate new license

### 4. Authentication Flow

The tests implement proper authentication progression:

```go
// 1. Initial registration (no auth required)
regData := struct {
    PublicKey string `json:"key"`
    InstallID string `json:"install_id"`
    FcmToken  string `json:"fcm_token"`
    Tos       string `json:"tos"`
    Model     string `json:"model"`
    Type      string `json:"type"`
    Locale    string `json:"locale"`
}{
    publicKey.String(),
    "",                    // Empty for non-mobile clients
    "",                    // Empty for non-mobile clients
    util.GetTimestamp(),   // Current timestamp for ToS acceptance
    "PC",                  // Device model
    "Android",             // Client type (matches mobile app)
    "en_US",               // Locale
}

// 2. Extract authentication credentials
deviceId := regResp["id"].(string)
accessToken := regResp["token"].(string)
initialLicenseKey := regResp["account"].(map[string]interface{})["license"].(string)

// 3. Set Bearer token for subsequent requests
defaultHeaders["Authorization"] = fmt.Sprintf("Bearer %s", accessToken)
```

**Authentication Progression:**
1. **Anonymous Registration**: Initial device creation without authentication
2. **Token Extraction**: Parse registration response for access token and device ID
3. **Authenticated Requests**: All subsequent API calls use Bearer token authentication
4. **License Management**: Handle license key updates and regeneration

## Integration with Optic

### Optic Workflow Integration

The tests integrate with Optic through the `optic.yml` configuration:

```yaml
name: wgcf
tasks:
  start:
    command: go run ./api_tests
    baseUrl: http://localhost:8889
```

**Integration Points:**
- **Test Execution**: Optic runs `go run ./api_tests` to execute the test suite
- **Proxy Setup**: Optic proxy intercepts HTTP traffic on `localhost:8889`
- **Specification Generation**: Captured interactions generate OpenAPI 3.0 specifications
- **Documentation Output**: Generated specs are saved to `.optic/generated/openapi.json`

### API Capture Process

1. **Proxy Startup**: Optic starts HTTP proxy server on configured port
2. **Test Execution**: Tests run and make API calls through the proxy
3. **Traffic Capture**: Proxy records all HTTP request/response pairs
4. **Schema Inference**: Optic analyzes captured data to infer API schemas
5. **Specification Generation**: Creates OpenAPI 3.0 specification with:
   - Path definitions for all tested endpoints
   - Request/response schema models
   - Parameter definitions and constraints
   - Authentication requirements
   - Error response patterns

## Test Execution and Configuration

### Running Tests

**Direct Execution:**
```bash
go run ./api_tests
```

**Via Optic (Recommended):**
```bash
api start  # Requires Optic CLI installed
```

**Docker Environment:**
```bash
docker run --rm -it -v $(pwd):/workspace optic/cli:latest start
```

### Prerequisites

1. **Optic CLI**: Install from https://github.com/opticdev/optic
2. **Go Runtime**: Go 1.23+ required for module support
3. **Network Access**: Tests require internet connectivity to reach Cloudflare API
4. **Valid Credentials**: Tests generate temporary credentials but need API access

### Configuration Options

**Environment Variables:**
- `WGCF_API_URL`: Override default Cloudflare API URL
- `WGCF_DEBUG`: Enable verbose logging for troubleshooting
- `OPTIC_PROXY_PORT`: Customize Optic proxy port (default: 8889)

**Test Customization:**
- Modify `defaultHeaders` for different client behavior
- Adjust `testConfig.InternetCheckTimeout` for network conditions
- Update test definitions to cover additional endpoints

## Adding New API Endpoint Tests

### Step-by-Step Process

1. **Identify New Endpoints**: Determine API endpoints that need documentation coverage
2. **Create Test Definition**: Add to the `tests` slice in `main.go`
3. **Handle Authentication**: Ensure proper auth token setup if required
4. **Test Request/Response**: Verify endpoint behavior and data structures
5. **Regenerate Documentation**: Run full generation pipeline

### Test Definition Template

```go
{
    "descriptive test name",           // Human-readable description
    requestBodyStruct,                 // Go struct for request payload (nil for GET)
    "endpoint/path/with/{variables}",  // API endpoint path
    "HTTP_METHOD",                     // GET, POST, PATCH, PUT, DELETE
}
```

### Example: Adding a New Endpoint

```go
// Add to the tests slice in main.go
{
    "get device statistics",
    nil,                                              // No request body for GET
    fmt.Sprintf("reg/%s/stats", deviceId),           // Endpoint with device ID
    "GET",                                           // HTTP method
},
{
    "update device location",
    struct {
        Latitude  float64 `json:"lat"`
        Longitude float64 `json:"lon"`
        Timezone  string  `json:"tz"`
    }{
        37.7749,                                     // San Francisco latitude
        -122.4194,                                   // San Francisco longitude
        "America/Los_Angeles",                       // Timezone
    },
    fmt.Sprintf("reg/%s/location", deviceId),        // Location update endpoint
    "PUT",                                           // Update method
},
```

### Authentication Considerations

**Anonymous Endpoints:**
- No additional setup required
- Examples: registration, client config

**Authenticated Endpoints:**
- Ensure `defaultHeaders["Authorization"]` is set
- Requires valid device registration first
- Examples: device updates, account management

**Administrative Endpoints:**
- May require elevated privileges
- Consider separate admin authentication flow
- Examples: license generation, account deletion

## Error Handling and Validation

### Response Validation

The custom transport implements strict response validation:

```go
if response.StatusCode != 200 {
    return nil, errors.New(fmt.Sprintf("bad code: %d", response.StatusCode))
}
```

**Validation Rules:**
- **Success Only**: Only HTTP 200 responses are considered successful
- **Error Propagation**: Non-200 responses generate test failures
- **Documentation Impact**: Failed requests are not included in generated specifications

### Error Scenarios

**Common Test Failures:**
1. **Network Connectivity**: Internet access required for Cloudflare API
2. **API Changes**: Cloudflare API modifications breaking compatibility
3. **Authentication Issues**: Invalid tokens or expired credentials
4. **Rate Limiting**: Too many requests in short timeframe
5. **TLS Configuration**: Incorrect TLS settings causing 403 errors

**Debugging Strategies:**
- Enable `DebugPrint: true` in test configuration
- Check Optic proxy logs for HTTP traffic details
- Verify network connectivity and firewall settings
- Validate API credentials and token format

## Integration with Generation Pipeline

### Full Generation Workflow

The API tests are part of a larger documentation generation pipeline:

1. **Test Execution**: `api_tests/main.go` runs and captures API interactions
2. **Specification Generation**: Optic generates `openapi.json` from captured data
3. **Specification Formatting**: `spec_format/main.go` cleans and formats the spec
4. **Client Generation**: OpenAPI Generator creates Go client code
5. **Documentation Updates**: Generated docs are committed to repository

### Pipeline Integration Points

**Shell Script Coordination** (`generate-api.sh`):
```bash
# 1. Generate OpenAPI spec from API exploration
api generate:oas --json
mv ".optic/generated/openapi.json" "openapi-spec.json"

# 2. Format the spec
go run "spec_format/main.go"

# 3. Generate Go client
openapi-generator-cli generate -i "openapi-spec.json" -g go -o "openapi"
```

**Dependencies:**
- Tests must complete successfully for spec generation
- Generated spec feeds into formatting and client generation
- Failed tests result in incomplete or invalid documentation

### Continuous Integration

**Automated Documentation Updates:**
- CI pipeline runs API tests on code changes
- Generated specifications are compared for breaking changes
- Documentation updates are automatically committed
- Client code regeneration ensures API compatibility

**Quality Gates:**
- All tests must pass for documentation generation
- Generated specifications must validate against OpenAPI 3.0 schema
- Client code must compile without errors
- Integration tests verify generated client functionality

## Best Practices

### Test Design Principles

1. **Comprehensive Coverage**: Test all public API endpoints
2. **Realistic Data**: Use representative request/response payloads
3. **Authentication Flows**: Cover both anonymous and authenticated scenarios
4. **Error Cases**: Include tests for common error conditions
5. **Idempotency**: Tests should not leave persistent state changes

### Maintenance Guidelines

1. **Regular Updates**: Update tests when API changes
2. **Documentation Sync**: Ensure tests match actual API behavior
3. **Performance**: Keep test execution time reasonable
4. **Reliability**: Tests should be deterministic and stable
5. **Security**: Never commit real credentials or sensitive data

### Troubleshooting Tips

**Common Issues:**
- **Port Conflicts**: Ensure Optic proxy port (8889) is available
- **Network Issues**: Verify internet connectivity and DNS resolution
- **Authentication**: Check token format and expiration
- **API Changes**: Validate endpoints against current Cloudflare API

**Debug Workflow:**
1. Enable debug logging in test configuration
2. Run tests individually to isolate issues
3. Check Optic proxy logs for HTTP traffic
4. Validate API responses manually with curl/Postman
5. Compare generated spec against expected schema

## Conclusion

The `api_tests/` directory provides a sophisticated API documentation generation system that captures real API behavior and produces accurate OpenAPI specifications. The integration with Optic enables automated documentation that stays synchronized with actual API implementation, while the comprehensive test coverage ensures reliable documentation generation.

The system balances automation with accuracy, providing both developers and API consumers with up-to-date, reliable documentation generated from actual API interactions rather than manually maintained specifications.