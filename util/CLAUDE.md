# Util Package Documentation

The `util` package provides essential utility functions for error handling, data manipulation, random generation, and type conversion used throughout the wgcf application.

## Available Utility Functions

### Error Handling

#### `GetErrorMessage(err error) string`

Formats error messages with full stack traces for debugging.

```go
// Usage example from main.go
if err := cmd.Execute(); err != nil {
    log.Fatal(util.GetErrorMessage(err))
}
```

**Purpose**: Provides detailed error information including stack traces using the `pkg/errors` package formatting.

### Random Generation

#### `RandomHexString(count int) string`

Generates a cryptographically secure random hexadecimal string.

```go
// Usage example from shared.go
deviceName += util.RandomHexString(3)
// Produces: "A1B2C3" (6 characters for 3 bytes)
```

**Parameters**:
- `count`: Number of random bytes to generate (output will be 2x this length in hex)

**Behavior**:
- Uses `crypto/rand` for secure random generation
- Panics if random generation fails (critical system error)
- Returns uppercase hexadecimal string

### Timestamp Generation

#### `GetTimestamp() string`

Returns current timestamp in RFC3339Nano format.

```go
// Usage example from api.go
timestamp := util.GetTimestamp()
// Produces: "2020-04-11T16:37:06.498123456+03:00"
```

**Internal Function**: `getTimestamp(t time.Time) string`
- Takes a specific time for testing purposes
- Uses `time.RFC3339Nano` format for maximum precision

### Type Conversion

#### `Restructure(source interface{}, dest interface{}) error`

Converts between different struct types using YAML serialization as an intermediate format.

```go
// Usage example from api.go
if err := util.Restructure(&result, &castResult); err != nil {
    return err
}
```

**How it works**:
1. Marshals source object to YAML
2. Unmarshals YAML to destination object
3. Provides more reliable conversion than reflection-based approaches

**Error Handling**: Returns wrapped errors with context using `pkg/errors.WithMessage()`.

## Error Handling Patterns

The util package follows consistent error handling patterns used throughout the application:

### Error Wrapping

Uses `github.com/pkg/errors` for error context:

```go
if err := yaml.Marshal(source); err != nil {
    return errors.WithMessage(err, "marshal")
}
```

### Error Formatting

Provides detailed error messages with stack traces:

```go
fmt.Sprintf("%+v", err) // Full stack trace formatting
```

### Standard Error Types

The application uses both standard Go errors and wrapped errors:

- `errors.New()` for simple error messages
- `errors.WithMessage()` for adding context to existing errors

## Testing Approach

### Test Coverage

The package includes unit tests in `util_test.go`:

```go
func TestGetTimestamp(t *testing.T) {
    expectedFormat := "2020-04-11T16:37:06.498+03:00"
    testTime := time.Date(2020, 04, 11, 16, 37, 06, 498*1000000, time.FixedZone("+3", 3*60*60))
    testFormat := getTimestamp(testTime)
    if testFormat != expectedFormat {
        t.Error("Invalid timestamp")
    }
}
```

### Testing Strategy

- **Deterministic testing**: Uses specific time values for consistent results
- **Format validation**: Ensures timestamp format compliance with RFC3339Nano
- **Timezone handling**: Tests with specific timezone offsets

### Running Tests

```bash
go test -v ./util/
```

## Usage Patterns Throughout Application

### Error Handling in Commands

All CLI commands use consistent error handling:

```go
if err := someOperation(); err != nil {
    log.Fatal(util.GetErrorMessage(err))
}
```

### API Timestamping

API requests include timestamps for tracking:

```go
timestamp := util.GetTimestamp()
// Used in API request headers or logging
```

### Device Name Generation

Random components are added to device names:

```go
if deviceName == "" {
    deviceName += util.RandomHexString(3)
}
```

### Type Conversion in APIs

Converting between API response types:

```go
var castResult TargetType
if err := util.Restructure(&apiResponse, &castResult); err != nil {
    return err
}
```

## Adding New Utilities

When adding new utility functions, follow these patterns:

### 1. Function Naming

Use clear, descriptive names that indicate the function's purpose:

```go
func GenerateSecureToken(length int) string
func ValidateEmailFormat(email string) bool
func ConvertBytesToHuman(bytes int64) string
```

### 2. Error Handling

Always use the established error handling patterns:

```go
func NewUtilityFunction(input string) (string, error) {
    if input == "" {
        return "", errors.New("input cannot be empty")
    }
    
    result, err := someOperation(input)
    if err != nil {
        return "", errors.WithMessage(err, "operation failed")
    }
    
    return result, nil
}
```

### 3. Add Tests

Create corresponding test functions:

```go
func TestNewUtilityFunction(t *testing.T) {
    result, err := NewUtilityFunction("test input")
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
    if result != "expected output" {
        t.Errorf("Expected 'expected output', got '%s'", result)
    }
}
```

### 4. Documentation

Add clear documentation comments:

```go
// NewUtilityFunction processes input and returns formatted output.
// Returns an error if input validation fails.
func NewUtilityFunction(input string) (string, error) {
    // Implementation here
}
```

## Dependencies

The util package depends on:

- `crypto/rand`: Secure random number generation
- `fmt`: String formatting
- `github.com/pkg/errors`: Enhanced error handling with stack traces
- `gopkg.in/yaml.v2`: YAML marshaling/unmarshaling for type conversion
- `time`: Timestamp generation

## File Structure

```
util/
├── util.go           # Main utility functions
├── util_test.go      # Unit tests
└── CLAUDE.md         # This documentation
```

## Best Practices

1. **Consistent Error Handling**: Always use `pkg/errors` for error wrapping and context
2. **Secure Randomness**: Use `crypto/rand` for any security-sensitive random generation
3. **Comprehensive Testing**: Test edge cases and error conditions
4. **Clear Documentation**: Document function behavior, parameters, and return values
5. **Type Safety**: Use type conversion utilities rather than unsafe casting
6. **Timezone Awareness**: Use RFC3339Nano format for timestamps to maintain timezone information

The util package serves as the foundational toolkit for common operations across the wgcf application, emphasizing reliability, security, and consistent error handling.