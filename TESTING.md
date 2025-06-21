# Testing Documentation

This document provides comprehensive information about the test suite for the score-checker application.

## Overview

The test suite follows Go testing best practices and uses only the standard library for maximum compatibility and simplicity. All tests are written with table-driven test patterns where appropriate and include comprehensive coverage of both success and failure scenarios.

## Test Structure

```
internal/
├── app/
│   └── app_test.go          # Application logic tests
├── config/
│   └── config_test.go       # Configuration loading tests
├── radarr/
│   └── client_test.go       # Radarr API client tests
├── sonarr/
│   └── client_test.go       # Sonarr API client tests
├── testhelpers/
│   └── testhelpers.go       # Test utilities and mock servers
└── types/
    └── types_test.go        # Type definitions tests
```

## Running Tests

### Basic Test Execution

```bash
# Run all tests
make test

# Run tests with verbose output
make test-verbose

# Run specific package tests
go test ./internal/app -v
go test ./internal/config -v
go test ./internal/sonarr -v
go test ./internal/radarr -v
go test ./internal/types -v
```

### Coverage

```bash
# Run tests with coverage percentages
make test-coverage

# Generate HTML coverage report
make test-coverage-html
```

Current coverage:
- **app**: 61.9% coverage
- **config**: 82.2% coverage  
- **radarr**: 80.0% coverage
- **sonarr**: 83.1% coverage
- **types**: 100% coverage (no statements)

### Benchmarks

```bash
# Run all benchmarks
make benchmark

# Run specific benchmarks
go test -bench=BenchmarkFindLowScoreEpisodes ./internal/app
go test -bench=BenchmarkFindLowScoreMovies ./internal/app
```

## Test Categories

### Unit Tests

#### Types Package (`internal/types/types_test.go`)
- **TestServiceConfig**: Validates service configuration structure
- **TestConfig**: Tests application configuration with multiple instances
- **TestSeries**: Validates series data structures
- **TestEpisode**: Tests episode data structures with file information
- **TestMovie**: Validates movie data structures
- **TestCommand**: Tests API command request/response structures

#### Config Package (`internal/config/config_test.go`)
- **TestInit**: Validates configuration initialization and defaults
- **TestLoadWithDefaults**: Tests default configuration loading
- **TestLoadWithEnvironmentVariables**: Tests environment variable configuration
- **TestLoadWithViperConfig**: Tests YAML configuration file loading
- **TestLoadWithDefaultInstanceNames**: Tests automatic instance naming
- **TestLoadEmptyInstanceArrays**: Tests handling of empty instance arrays

#### Sonarr Client (`internal/sonarr/client_test.go`)
- **TestNewClient**: Validates client initialization
- **TestGetSeries**: Tests series retrieval with various response scenarios
- **TestGetEpisodes**: Tests episode retrieval with file information
- **TestTriggerEpisodeSearch**: Tests search command triggering
- **TestMakeRequest**: Tests HTTP request handling and error scenarios

#### Radarr Client (`internal/radarr/client_test.go`)
- **TestNewClient**: Validates client initialization
- **TestGetMovies**: Tests movie retrieval with various response scenarios
- **TestTriggerMovieSearch**: Tests movie search command triggering
- **TestMakeRequest**: Tests HTTP request handling and error scenarios

#### App Package (`internal/app/app_test.go`)
- **TestFindLowScoreEpisodes**: Tests episode processing logic with various configurations
- **TestFindLowScoreMovies**: Tests movie processing logic with batch limiting
- **TestPrintLowScoreEpisodes**: Tests console output formatting for episodes
- **TestPrintLowScoreMovies**: Tests console output formatting for movies

### Integration Tests

Currently limited due to the need for better dependency injection. The `TestRunOnceIntegration` test is skipped as it requires significant refactoring for proper testability.

### Test Helpers

The `testhelpers` package provides:

- **MockSonarrServer**: HTTP test server that simulates Sonarr API responses
- **MockRadarrServer**: HTTP test server that simulates Radarr API responses  
- **TestingInterface**: Interface allowing both `*testing.T` and `*testing.B` for shared test utilities
- **Test Data Factories**: Functions to create consistent test data across test suites

## Test Patterns

### Table-Driven Tests

Most tests use table-driven patterns for comprehensive scenario coverage:

```go
tests := []struct {
    name           string
    input          InputType
    expectedOutput OutputType
    expectError    bool
}{
    {
        name:           "success case",
        input:          validInput,
        expectedOutput: expectedOutput,
        expectError:    false,
    },
    {
        name:        "error case",
        input:       invalidInput,
        expectError: true,
    },
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        // Test implementation
    })
}
```

### Mock HTTP Servers

API client tests use `httptest.Server` to simulate real API interactions:

```go
server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Verify request headers, method, and body
    // Return appropriate test response
}))
defer server.Close()

client := NewClient(types.ServiceConfig{
    BaseURL: server.URL,
    APIKey:  "test-key",
})
```

### Output Capture

Tests that verify console output capture stdout during execution:

```go
oldStdout := os.Stdout
r, w, _ := os.Pipe()
os.Stdout = w

// Function that produces output
printFunction()

w.Close()
os.Stdout = oldStdout

// Read and verify captured output
```

## Best Practices Implemented

1. **Standard Library Only**: No external testing dependencies for maximum compatibility
2. **Comprehensive Coverage**: Tests cover both success and failure scenarios
3. **Clear Test Names**: Descriptive test names that explain the scenario being tested
4. **Isolated Tests**: Each test is independent and can run in any order
5. **Mock External Dependencies**: HTTP APIs are mocked using standard library tools
6. **Table-Driven Tests**: Consistent pattern for testing multiple scenarios
7. **Benchmark Tests**: Performance tests for critical application paths
8. **Test Helpers**: Reusable utilities to reduce test code duplication

## CI/CD Integration

The test suite is designed to integrate easily with CI/CD pipelines:

```bash
# Quality gate commands
make fmt      # Code formatting
make vet      # Go vet static analysis  
make test     # Run test suite
make lint     # Run linter (requires golangci-lint)
```

## Extending Tests

When adding new functionality:

1. **Add unit tests** for new functions and methods
2. **Update existing tests** if changing behavior
3. **Add integration tests** for new API endpoints or major features
4. **Update test helpers** if new mock functionality is needed
5. **Maintain coverage** above 75% for critical packages

## Known Limitations

1. **Integration Testing**: Limited due to tight coupling between config and application logic
2. **Error Testing**: Some error scenarios require significant setup complexity
3. **Concurrency Testing**: Current tests don't cover concurrent access patterns

## Future Improvements

1. **Dependency Injection**: Refactor application for better testability
2. **Contract Testing**: Add tests for API contract compliance
3. **Property-Based Testing**: Consider adding property-based tests for complex logic
4. **Integration Tests**: Add comprehensive integration test suite
5. **Performance Testing**: Expand benchmark coverage for all critical paths