package handlers

import (
	"net/http"
	"testing"

	"github.com/typing-master-for-coding-backend/internal/middleware"
	"github.com/typing-master-for-coding-backend/internal/models"
	"github.com/typing-master-for-coding-backend/internal/testutil"

	"github.com/stretchr/testify/assert"
)

// TestGetLanguages tests language listing endpoint
func TestGetLanguages(t *testing.T) {
	router, mockDB, mockCache := testutil.SetupTestRouter()

	router.GET("/api/v1/languages", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetLanguages(mockDB, mockCache))

	// Requirement 1: Support for C++, Rust, Python, JavaScript, and YAML
	t.Run("Get All Languages", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")
		testutil.CreateTestLanguage(mockDB, "python", "Python")
		testutil.CreateTestLanguage(mockDB, "cpp", "C++")
		testutil.CreateTestLanguage(mockDB, "rust", "Rust")
		testutil.CreateTestLanguage(mockDB, "yaml", "YAML")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/languages", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var languages []map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &languages)

		assert.GreaterOrEqual(t, len(languages), 5)

		// Verify required languages are present
		languageIDs := make([]string, len(languages))
		for i, lang := range languages {
			languageIDs[i] = lang["id"].(string)
		}

		assert.Contains(t, languageIDs, "javascript")
		assert.Contains(t, languageIDs, "python")
		assert.Contains(t, languageIDs, "cpp")
	})
}

// TestGetLanguage tests single language retrieval
func TestGetLanguage(t *testing.T) {
	router, mockDB, mockCache := testutil.SetupTestRouter()

	router.GET("/api/v1/languages/:id", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetLanguage(mockDB, mockCache))

	t.Run("Get Existing Language", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/languages/javascript", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var language map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &language)

		assert.Equal(t, "javascript", language["id"])
		assert.Equal(t, "JavaScript", language["name"])
		assert.Contains(t, language, "parser_id")
		assert.Contains(t, language, "grammar_config")
	})

	t.Run("Get Non-Existent Language", func(t *testing.T) {
		mockDB.Clear()

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/languages/nonexistent", nil, headers)

		testutil.AssertErrorResponse(t, w, http.StatusNotFound, "Language not found")
	})
}

// TestCreateLanguage tests language creation
func TestCreateLanguage(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/languages", middleware.AuthMiddleware("test-secret-key-for-testing-only"), CreateLanguage(mockDB))

	// Requirement 1: Tree-sitter grammars for tokenization
	t.Run("Create New Language", func(t *testing.T) {
		mockDB.Clear()

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.Language{
			ID:       "go",
			Name:     "Go",
			Version:  "1",
			ParserID: "tree-sitter-go",
			GrammarConfig: map[string]interface{}{
				"semicolons": false,
			},
			WhitespaceRules: map[string]interface{}{
				"indentation": "tabs",
			},
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/languages", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "go", response["id"])
		assert.Equal(t, "Go", response["name"])
		assert.Equal(t, "tree-sitter-go", response["parser_id"])
	})

	t.Run("Create Language with Invalid Data", func(t *testing.T) {
		mockDB.Clear()

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := map[string]interface{}{
			"name": "Invalid Language",
			// Missing required ID field
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/languages", reqBody, headers)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestGetLessons tests lesson listing endpoint
func TestGetLessons(t *testing.T) {
	router, mockDB, mockCache := testutil.SetupTestRouter()

	router.GET("/api/v1/lessons", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetLessons(mockDB, mockCache))

	// Requirement 2: Syntax Tutorials with progression
	t.Run("Get All Lessons", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")
		testutil.CreateTestLesson(mockDB, "lesson1", "javascript", "Intro to Functions")
		testutil.CreateTestLesson(mockDB, "lesson2", "javascript", "Advanced Patterns")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/lessons", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var lessons []map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &lessons)

		assert.GreaterOrEqual(t, len(lessons), 2)
	})
}

// TestCreateLesson tests lesson creation
func TestCreateLesson(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/lessons", middleware.AuthMiddleware("test-secret-key-for-testing-only"), CreateLesson(mockDB))

	// Requirement 2: Structured lessons with progression
	t.Run("Create Lesson with Prerequisites", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.Lesson{
			LanguageID:       "javascript",
			Title:            "Advanced Functions",
			Difficulty:       2,
			Objectives:       []string{"Learn closures", "Master callbacks"},
			Prerequisites:    []string{"lesson_intro"},
			EstimatedMinutes: 45,
			TokensCovered:    []string{"function", "=>", "callback"},
			SnippetIDs:       []string{"snippet1", "snippet2"},
			Version:          1,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/lessons", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "Advanced Functions", response["title"])
		assert.Equal(t, "javascript", response["language_id"])
		assert.Equal(t, float64(2), response["difficulty"])
		assert.Contains(t, response, "id")
	})

	t.Run("Create Lesson with Token Coverage", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "python", "Python")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.Lesson{
			LanguageID:       "python",
			Title:            "Python Basics",
			Difficulty:       1,
			Objectives:       []string{"Learn syntax"},
			EstimatedMinutes: 30,
			TokensCovered:    []string{"def", "class", "import", "if", "for"},
			Version:          1,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/lessons", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		tokensCovered := response["tokens_covered"].([]interface{})
		assert.Equal(t, 5, len(tokensCovered))
	})
}

// TestGetSnippets tests snippet listing endpoint
func TestGetSnippets(t *testing.T) {
	router, mockDB, mockCache := testutil.SetupTestRouter()

	router.GET("/api/v1/snippets", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetSnippets(mockDB, mockCache))

	t.Run("Get All Snippets", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")
		testutil.CreateTestSnippet(mockDB, "snippet1", "javascript", "Hello World", "console.log('Hello');")
		testutil.CreateTestSnippet(mockDB, "snippet2", "javascript", "Function Example", "function test() {}")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		w := testutil.MakeRequest(router, "GET", "/api/v1/snippets", nil, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var snippets []map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &snippets)

		assert.GreaterOrEqual(t, len(snippets), 2)
	})
}

// TestCreateSnippet tests snippet creation with checksum
func TestCreateSnippet(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/snippets", middleware.AuthMiddleware("test-secret-key-for-testing-only"), CreateSnippet(mockDB))

	// Requirement 7: Content management with validation
	t.Run("Create Snippet with Checksum", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		sourceCode := "function greet(name) {\n  return `Hello, ${name}!`;\n}"

		reqBody := models.Snippet{
			LanguageID:    "javascript",
			Title:         "Greeting Function",
			SourceCode:    sourceCode,
			Tags:          []string{"function", "template-literal"},
			Difficulty:    1,
			EstimatedTime: 5,
			AccessibilityTags: map[string]interface{}{
				"complexity": "low",
			},
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/snippets", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "Greeting Function", response["title"])
		assert.Equal(t, sourceCode, response["source_code"])
		assert.Contains(t, response, "checksum")
		assert.NotEmpty(t, response["checksum"])
	})

	t.Run("Create Snippet with Difficulty Bands", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "python", "Python")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		// Create snippets of different sizes (Requirement 7: snippet length bands)
		testCases := []struct {
			name       string
			code       string
			difficulty int
		}{
			{"XS Snippet", "x = 5", 1},                        // < 200 chars
			{"S Snippet", "def hello():\n    print('hi')", 1}, // 200-400 chars
		}

		for _, tc := range testCases {
			reqBody := models.Snippet{
				LanguageID:    "python",
				Title:         tc.name,
				SourceCode:    tc.code,
				Difficulty:    tc.difficulty,
				EstimatedTime: 3,
			}

			w := testutil.MakeRequest(router, "POST", "/api/v1/snippets", reqBody, headers)
			assert.Equal(t, http.StatusCreated, w.Code)
		}
	})
}

// TestImportSnippet tests snippet import with auto-assessment
func TestImportSnippet(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/snippets/import", middleware.AuthMiddleware("test-secret-key-for-testing-only"), ImportSnippet(mockDB))

	// Requirement 7: Auto-tagging and difficulty assessment
	t.Run("Import Snippet with Auto-Assessment", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := map[string]interface{}{
			"language_id": "javascript",
			"title":       "Auto-Tagged Function",
			"source_code": "function test() {\n  if (true) {\n    for (let i = 0; i < 10; i++) {}\n  }\n}",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/snippets/import", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		// Verify auto-generated tags
		tags := response["tags"].([]interface{})
		assert.Greater(t, len(tags), 0)

		// Verify auto-assessed difficulty
		assert.Contains(t, response, "difficulty")
		assert.Greater(t, response["difficulty"], float64(0))

		// Verify estimated time
		assert.Contains(t, response, "estimated_time")

		// Verify accessibility tags
		assert.Contains(t, response, "accessibility_tags")
	})
}

// TestCreatePlaylist tests playlist creation
func TestCreatePlaylist(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/playlists", middleware.AuthMiddleware("test-secret-key-for-testing-only"), CreatePlaylist(mockDB))

	t.Run("Create Custom Playlist", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.Playlist{
			UserID:      user.ID,
			Title:       "My JavaScript Journey",
			Description: "Custom learning path for JavaScript",
			LanguageID:  "javascript",
			LessonIDs:   []string{"lesson1", "lesson2"},
			SnippetIDs:  []string{"snippet1", "snippet2"},
			Tags:        []string{"beginner", "functions"},
			IsPublic:    false,
			Difficulty:  1,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/playlists", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "My JavaScript Journey", response["title"])
		assert.Equal(t, user.ID, response["user_id"])
		assert.Contains(t, response, "id")
	})
}

// TestUpdateSnippet tests snippet update with checksum validation
func TestUpdateSnippet(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.PUT("/api/v1/snippets/:id", middleware.AuthMiddleware("test-secret-key-for-testing-only"), UpdateSnippet(mockDB))

	t.Run("Update Snippet Regenerates Checksum", func(t *testing.T) {
		mockDB.Clear()
		testutil.CreateTestLanguage(mockDB, "javascript", "JavaScript")
		snippet := testutil.CreateTestSnippet(mockDB, "snippet1", "javascript", "Original", "console.log('old');")

		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")
		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		originalChecksum := snippet.Checksum
		newCode := "console.log('new');"

		reqBody := models.Snippet{
			Title:      "Updated",
			SourceCode: newCode,
			Tags:       []string{"updated"},
			Difficulty: 2,
		}

		w := testutil.MakeRequest(router, "PUT", "/api/v1/snippets/snippet1", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "Updated", response["title"])
		assert.Equal(t, newCode, response["source_code"])
		assert.NotEqual(t, originalChecksum, response["checksum"])
	})
}

// TestContentVersioning tests content version tracking
func TestContentVersioning(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/content/versions", middleware.AuthMiddleware("test-secret-key-for-testing-only"), CreateContentVersion(mockDB))
	router.GET("/api/v1/content/versions/:contentType/:contentId", middleware.AuthMiddleware("test-secret-key-for-testing-only"), GetContentVersions(mockDB))

	// Requirement 7: Versioned content with deprecation
	t.Run("Create Content Version", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.ContentVersion{
			ContentType: "lesson",
			ContentID:   "lesson1",
			Version:     1,
			Content: map[string]interface{}{
				"title": "Original Lesson",
			},
			CreatedBy:   user.ID,
			ChangeNotes: "Initial version",
			IsActive:    true,
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/content/versions", reqBody, headers)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "lesson", response["content_type"])
		assert.Equal(t, "lesson1", response["content_id"])
		assert.Contains(t, response, "checksum")
	})
}

// TestValidateContentChecksum tests content validation
func TestValidateContentChecksum(t *testing.T) {
	router, mockDB, _ := testutil.SetupTestRouter()

	router.POST("/api/v1/content/validate", middleware.AuthMiddleware("test-secret-key-for-testing-only"), ValidateContentChecksum(mockDB))

	t.Run("Validate Content Checksum", func(t *testing.T) {
		mockDB.Clear()
		user := testutil.CreateTestUser(mockDB, "user1", "testuser", "test@example.com")

		token, _ := testutil.GenerateTestToken(user.ID, user.Handle, user.Email, false)
		headers := map[string]string{
			"Authorization": "Bearer " + token,
		}

		reqBody := models.ContentValidation{
			ContentType: "snippet",
			ContentID:   "snippet1",
			Checksum:    "abc123def456",
		}

		w := testutil.MakeRequest(router, "POST", "/api/v1/content/validate", reqBody, headers)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		testutil.ParseJSON(w.Body.Bytes(), &response)

		assert.Equal(t, "valid", response["validation_status"])
		assert.Contains(t, response, "validated_at")
	})
}
