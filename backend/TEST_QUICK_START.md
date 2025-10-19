# Test Quick Start Guide

## 🚀 Running Tests

### Windows (PowerShell)

```powershell
# Run all tests
.\run-tests.ps1

# Run with verbose output and coverage
.\run-tests.ps1 -Verbose -Coverage

# Run specific test
.\run-tests.ps1 -Run "TestRegister" -Verbose

# Run with race detection
.\run-tests.ps1 -Race -Verbose
```

### Linux/Mac (Bash)

```bash
# Make script executable (first time only)
chmod +x run-tests.sh

# Run all tests
./run-tests.sh

# Run with verbose output and coverage
./run-tests.sh --verbose --coverage

# Run specific test
./run-tests.sh --run TestRegister --verbose

# Run with race detection
./run-tests.sh --race --verbose
```

## 📊 Test Coverage

```powershell
# Generate coverage report
.\run-tests.ps1 -Coverage

# View coverage in browser
start coverage.html  # Windows
open coverage.html   # Mac
xdg-open coverage.html  # Linux
```

## 🧪 Test Files Overview

| File | Description | Test Count |
|------|-------------|------------|
| `auth_test.go` | Authentication & user management | 20+ tests |
| `privacy_test.go` | GDPR compliance & privacy | 15+ tests |
| `content_test.go` | Content management (languages, lessons, snippets) | 25+ tests |
| `progression_test.go` | Lesson progression & tracking | 15+ tests |
| `session_test.go` | Sessions, results, public endpoints | 15+ tests |

**Total: 90+ comprehensive unit tests**

## ✅ Requirements Validation

### Requirement 1: Multi-Language Support
- ✅ C++, Rust, Python, JavaScript, YAML
- ✅ Tree-sitter grammar configuration
- ✅ Language CRUD operations

### Requirement 2: Practice Modes
- ✅ Lesson progression (Intro → Core → Idioms → Advanced → Review)
- ✅ Prerequisites system
- ✅ Progress tracking

### Requirement 3: Metrics & Scoring
- ✅ Progression summary
- ⏳ Session metrics (endpoints not yet implemented)

### Requirement 4: Leaderboards
- ⏳ Leaderboard structure (not yet implemented)

### Requirement 5: Customization
- ✅ Keyboard layouts (QWERTY, Dvorak, etc.)
- ✅ User settings & preferences

### Requirement 6: Privacy & GDPR
- ✅ Anonymous mode
- ✅ Data export (JSON/CSV)
- ✅ Right to deletion
- ✅ Consent tracking

### Requirement 7: Content Management
- ✅ Versioning & validation
- ✅ Auto-assessment & tagging
- ✅ Checksum generation

## 🔍 Test Examples

### Authentication Test
```go
t.Run("Successful Registration", func(t *testing.T) {
    mockDB.Clear()
    
    reqBody := models.RegisterRequest{
        Handle:   "testuser",
        Email:    "test@example.com",
        Password: "password123",
    }
    
    w := testutil.MakeRequest(router, "POST", "/api/v1/auth/register", reqBody, nil)
    
    assert.Equal(t, http.StatusCreated, w.Code)
})
```

### Privacy Test
```go
t.Run("Export User Data as JSON", func(t *testing.T) {
    mockDB.Clear()
    user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
    
    token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
    headers := map[string]string{
        "Authorization": "Bearer " + token,
    }
    
    w := testutil.MakeRequest(router, "GET", "/api/v1/privacy/export?format=json", nil, headers)
    
    assert.Equal(t, http.StatusOK, w.Code)
})
```

## 📝 Common Commands

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Run specific package
go test ./internal/handlers

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...
```

## 🐛 Troubleshooting

### Issue: Tests fail with "User not found"
**Solution**: Ensure `mockDB.Clear()` is called at the start of each test

### Issue: Token authentication fails
**Solution**: Verify JWT secret matches and token is in Authorization header

### Issue: Mock database issues
**Solution**: Check entities are saved with correct keys and query filters match

## 📚 Documentation

- **Full Testing Guide**: See `TESTING.md`
- **Requirements**: See `.kiro/specs/typing-master-for-coding/requirements.md`
- **Design**: See `.kiro/specs/typing-master-for-coding/design.md`

## 🎯 Next Steps

1. Run tests to verify everything works
2. Review test coverage report
3. Add tests for any new endpoints
4. Maintain >80% code coverage

---

**Need Help?** Check the full `TESTING.md` documentation for detailed information.
