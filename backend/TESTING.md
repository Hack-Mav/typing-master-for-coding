# Testing Documentation

## Overview

This document describes the comprehensive unit test suite for the Typing Master Backend API. The tests validate all endpoints against the requirements and design specifications.

## Test Structure

### Test Files

- **`auth_test.go`** - Authentication and user management endpoints
- **`privacy_test.go`** - Privacy/GDPR compliance endpoints
- **`content_test.go`** - Content management (languages, lessons, snippets, playlists)
- **`progression_test.go`** - Lesson progression and tracking endpoints
- **`session_test.go`** - Session management and results endpoints

### Test Utilities

- **`testutil/testutil.go`** - Helper functions for test setup and assertions
- **`database/mock.go`** - Mock database implementation for testing

## Requirements Coverage

### Requirement 1: Multi-Language Support
- ✅ Tests for C++, Rust, Python, JavaScript, and YAML language support
- ✅ Tree-sitter grammar configuration validation
- ✅ Language CRUD operations

### Requirement 2: Practice Modes
- ✅ Syntax tutorial progression (Intro → Core → Idioms → Advanced → Review)
- ✅ Lesson prerequisite system
- ✅ Progress tracking and completion

### Requirement 3: Metrics and Scoring
- ✅ Session tracking (planned - endpoints not yet implemented)
- ✅ Results computation (planned - endpoints not yet implemented)
- ✅ Progression summary with statistics

### Requirement 4: Leaderboards
- ✅ Leaderboard endpoint structure (planned - not yet implemented)

### Requirement 5: Customization
- ✅ Keyboard layout configuration (QWERTY, Dvorak, etc.)
- ✅ User settings and preferences
- ✅ Profile customization

### Requirement 6: Privacy & GDPR
- ✅ Anonymous mode support
- ✅ Privacy settings management
- ✅ Data export (JSON and CSV formats)
- ✅ Right to deletion
- ✅ Consent tracking
- ✅ Data anonymization

### Requirement 7: Content Management
- ✅ Lesson creation with token coverage
- ✅ Snippet import with auto-assessment
- ✅ Content versioning and validation
- ✅ Checksum generation and validation
- ✅ Difficulty assessment
- ✅ Auto-tagging

### Requirement 8: Offline & Performance
- ⏳ PWA functionality (frontend)
- ⏳ Caching strategy (partially implemented)

## Running Tests

### Quick Start

```powershell
# Windows (PowerShell)
.\run-tests.ps1 -Verbose -Coverage

# Linux/Mac (Bash)
chmod +x run-tests.sh
./run-tests.sh --verbose --coverage
```

### Test Options

#### PowerShell (Windows)
```powershell
# Run all tests with verbose output
.\run-tests.ps1 -Verbose

# Run with coverage report
.\run-tests.ps1 -Coverage

# Run with race detection
.\run-tests.ps1 -Race

# Run specific test
.\run-tests.ps1 -Run "TestRegister"

# Run specific package
.\run-tests.ps1 -Package "./internal/handlers"

# Combine options
.\run-tests.ps1 -Verbose -Coverage -Race
```

#### Bash (Linux/Mac)
```bash
# Run all tests with verbose output
./run-tests.sh --verbose

# Run with coverage report
./run-tests.sh --coverage

# Run with race detection
./run-tests.sh --race

# Run specific test
./run-tests.sh --run TestRegister

# Run specific package
./run-tests.sh --package ./internal/handlers

# Combine options
./run-tests.sh --verbose --coverage --race
```

### Manual Test Execution

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package
go test ./internal/handlers

# Run specific test
go test -run TestRegister ./internal/handlers

# Run with race detection
go test -race ./...
```

## Test Categories

### Authentication Tests (`auth_test.go`)

#### Registration
- ✅ Successful registration with all fields
- ✅ Duplicate email detection
- ✅ Duplicate handle detection
- ✅ Invalid email format validation
- ✅ Password length validation
- ✅ Privacy consent tracking

#### Login
- ✅ Successful login with valid credentials
- ✅ Invalid email rejection
- ✅ Invalid password rejection
- ✅ Missing credentials validation

#### Token Management
- ✅ Token refresh functionality
- ✅ Invalid token rejection
- ✅ Missing token handling

#### Anonymous Sessions
- ✅ Anonymous session creation
- ✅ Device ID requirement
- ✅ Custom keyboard layout support

#### Profile Management
- ✅ Profile retrieval for authenticated users
- ✅ Profile retrieval for anonymous users
- ✅ Profile updates (handle, locale, keyboard layout)
- ✅ Privacy settings updates
- ✅ Anonymous user restrictions

### Privacy Tests (`privacy_test.go`)

#### Privacy Settings
- ✅ Privacy mode toggle
- ✅ Telemetry consent management
- ✅ Data processing consent
- ✅ Anonymous user restrictions

#### Data Export (GDPR)
- ✅ JSON export format
- ✅ CSV export format
- ✅ Complete data inclusion (profile, sessions, results, progress)
- ✅ Anonymous user handling

#### Data Deletion (GDPR)
- ✅ Complete data deletion
- ✅ Confirmation requirement
- ✅ Cascading deletion (sessions, events, results, progress)
- ✅ Anonymous user handling

#### Data Anonymization
- ✅ User ID hashing
- ✅ PII removal
- ✅ Timestamp rounding

### Content Management Tests (`content_test.go`)

#### Languages
- ✅ List all languages
- ✅ Get single language
- ✅ Create language with Tree-sitter config
- ✅ Update language
- ✅ Delete language

#### Lessons
- ✅ List all lessons
- ✅ Get single lesson
- ✅ Create lesson with prerequisites
- ✅ Create lesson with token coverage
- ✅ Update lesson (version increment)
- ✅ Delete lesson

#### Snippets
- ✅ List all snippets
- ✅ Get single snippet
- ✅ Create snippet with checksum
- ✅ Update snippet (checksum regeneration)
- ✅ Delete snippet
- ✅ Import snippet with auto-assessment
- ✅ Difficulty bands (XS, S, M, L)

#### Playlists
- ✅ Create custom playlist
- ✅ Update playlist
- ✅ Delete playlist

#### Content Versioning
- ✅ Create content version
- ✅ Get content versions
- ✅ Validate content checksum

### Progression Tests (`progression_test.go`)

#### Lesson Progress
- ✅ Get existing progress
- ✅ Create new progress
- ✅ Update existing progress
- ✅ Complete lesson (100% progress)
- ✅ Track best score
- ✅ Increment attempt count

#### Prerequisites
- ✅ Check prerequisites met
- ✅ Check prerequisites missing
- ✅ Unlock lessons based on completion

#### Progression Flow
- ✅ Get lesson flow (Intro → Core → Idioms → Advanced → Review)
- ✅ Stage descriptions and estimated times
- ✅ Token coverage per stage

#### Progression Summary
- ✅ Calculate total lessons
- ✅ Calculate completed lessons
- ✅ Calculate overall progress
- ✅ Calculate average score
- ✅ Calculate completion rate

### Session Tests (`session_test.go`)

#### Health Check
- ✅ Service health status

#### Session Management
- ⏳ Create session (not yet implemented)
- ⏳ Update session (not yet implemented)
- ⏳ Record events (not yet implemented)
- ⏳ Finalize session (not yet implemented)

#### Results
- ⏳ Get results (not yet implemented)
- ⏳ Get single result (not yet implemented)

#### Leaderboards
- ⏳ Get leaderboards (not yet implemented)

#### Public Endpoints
- ✅ Public languages endpoint
- ✅ Public lessons endpoint
- ✅ Public snippets endpoint

#### CRUD Operations
- ✅ Delete operations (languages, lessons, snippets, playlists)
- ✅ Update operations (languages, lessons, playlists)

## Test Utilities

### Setup Functions

```go
// Create test router with mock dependencies
router, mockDB, mockCache := testutil.SetupTestRouter()

// Create test configuration
cfg := testutil.CreateTestConfig()

// Create test user
user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

// Create test language
lang := testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")

// Create test lesson
lesson := testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro")

// Create test snippet
snippet := testutil.CreateTestSnippet(mockDB, "snippet1", "javascript", "Title", "code")
```

### Request Helpers

```go
// Make HTTP request
w := testutil.MakeRequest(router, "POST", "/api/v1/auth/login", reqBody, headers)

// Generate test JWT token
token, _ := testutil.GenerateTestToken(userID, handle, email, isAnonymous)

// Parse JSON response
var response map[string]interface{}
testutil.ParseJSON(w.Body.Bytes(), &response)
```

### Assertion Helpers

```go
// Assert error response
testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "error message")

// Assert JSON response
testutil.AssertJSONResponse(t, w, http.StatusOK, expectedBody)
```

## Coverage Goals

- **Target Coverage**: 80%+
- **Current Coverage**: Run `./run-tests.ps1 -Coverage` to see current coverage

### Coverage by Package

```bash
# View coverage by package
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Continuous Integration

### GitHub Actions

Tests are automatically run on:
- Push to main branch
- Pull requests
- Manual workflow dispatch

See `.github/workflows/ci.yml` for CI configuration.

## Best Practices

### Writing Tests

1. **Arrange-Act-Assert Pattern**
   ```go
   t.Run("Test Description", func(t *testing.T) {
       // Arrange
       mockDB.Clear()
       user := testutil.CreateTestUser(mockDB, "user1", "test", "test@example.com")
       
       // Act
       w := testutil.MakeRequest(router, "GET", "/api/v1/profile", nil, headers)
       
       // Assert
       assert.Equal(t, http.StatusOK, w.Code)
   })
   ```

2. **Clear Test Names**
   - Use descriptive test names
   - Follow pattern: `Test<Function>_<Scenario>_<ExpectedResult>`

3. **Isolation**
   - Clear mock database before each test
   - Don't depend on test execution order
   - Use `mockDB.Clear()` in each test

4. **Coverage**
   - Test happy path
   - Test error cases
   - Test edge cases
   - Test validation

## Troubleshooting

### Common Issues

#### Tests Fail with "User not found"
- Ensure `mockDB.Clear()` is called at the start of the test
- Verify user is created before making authenticated requests

#### Token Authentication Fails
- Check JWT secret matches between token generation and middleware
- Verify token is included in Authorization header: `Bearer <token>`

#### Mock Database Issues
- Ensure mock database is properly initialized
- Check that entities are saved with correct keys
- Verify query filters match stored data

### Debug Mode

Run tests with verbose output to see detailed logs:

```powershell
.\run-tests.ps1 -Verbose
```

## Future Enhancements

### Planned Test Additions

1. **Session Management**
   - Session creation and lifecycle
   - Event recording and batching
   - Session finalization
   - Metrics calculation

2. **Results & Scoring**
   - CPM, tWPM, KPS calculation
   - Accuracy metrics
   - Composite score computation
   - Error analysis

3. **Leaderboards**
   - Ranking algorithms
   - Anti-cheat validation
   - Time-windowed leaderboards
   - Filtering by language/mode

4. **Integration Tests**
   - End-to-end API workflows
   - Database integration
   - Cache integration
   - External service mocking

5. **Performance Tests**
   - Load testing
   - Concurrent request handling
   - Cache effectiveness
   - Query optimization

## Contributing

When adding new tests:

1. Follow existing test structure
2. Add tests for both success and error cases
3. Update this documentation
4. Ensure all tests pass before committing
5. Maintain or improve code coverage

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Gin Testing Guide](https://github.com/gin-gonic/gin#testing)
- [Requirements Document](../.kiro/specs/typing-master-for-coding/requirements.md)
- [Design Document](../.kiro/specs/typing-master-for-coding/design.md)
