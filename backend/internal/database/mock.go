package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/typing-master-for-coding-backend/internal/models"

	"cloud.google.com/go/datastore"
)

// MockDatastore provides an in-memory implementation for development
type MockDatastore struct {
	mu                 sync.RWMutex
	users              map[string]*models.User
	languages          map[string]*models.Language
	lessons            map[string]*models.Lesson
	snippets           map[string]*models.Snippet
	sessions           map[string]*models.Session
	events             map[string][]*models.SessionEvent
	results            map[string]*models.Result
	playlists          map[string]*models.Playlist
	lessonProgress     map[string]*models.LessonProgress
	contentVersions    map[string]*models.ContentVersion
	contentValidations map[string]*models.ContentValidation
}

// MockDatastoreClient is an alias for testing compatibility
type MockDatastoreClient = MockDatastore

func NewMockDatastore() *MockDatastore {
	mock := &MockDatastore{
		users:              make(map[string]*models.User),
		languages:          make(map[string]*models.Language),
		lessons:            make(map[string]*models.Lesson),
		snippets:           make(map[string]*models.Snippet),
		sessions:           make(map[string]*models.Session),
		events:             make(map[string][]*models.SessionEvent),
		results:            make(map[string]*models.Result),
		playlists:          make(map[string]*models.Playlist),
		lessonProgress:     make(map[string]*models.LessonProgress),
		contentVersions:    make(map[string]*models.ContentVersion),
		contentValidations: make(map[string]*models.ContentValidation),
	}

	// Initialize with some sample data
	mock.initSampleData()

	return mock
}

// NewMockDatastoreClient creates a new mock datastore client for testing
func NewMockDatastoreClient() *MockDatastoreClient {
	return NewMockDatastore()
}

func (m *MockDatastore) initSampleData() {
	// Add sample languages
	languages := []*models.Language{
		{
			ID:              "javascript",
			Name:            "JavaScript",
			Version:         1,
			ParserID:        "tree-sitter-javascript",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "python",
			Name:            "Python",
			Version:         1,
			ParserID:        "tree-sitter-python",
			GrammarConfig:   map[string]interface{}{"indentation": "spaces"},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "cpp",
			Name:            "C++",
			Version:         1,
			ParserID:        "tree-sitter-cpp",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "rust",
			Name:            "Rust",
			Version:         1,
			ParserID:        "tree-sitter-rust",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "yaml",
			Name:            "YAML",
			Version:         1,
			ParserID:        "tree-sitter-yaml",
			GrammarConfig:   map[string]interface{}{"indentation": "spaces"},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
	}

	for _, lang := range languages {
		m.languages[lang.ID] = lang
	}

	// Add sample lessons
	lessons := []*models.Lesson{
		{
			ID:               "lesson1",
			LanguageID:       "javascript",
			Title:            "Intro to Functions",
			Difficulty:       1,
			Objectives:       []string{"Learn JavaScript functions"},
			Prerequisites:    []string{},
			EstimatedMinutes: 30,
			TokensCovered:    []string{"function", "return"},
			SnippetIDs:       []string{},
			Version:          1,
			CreatedBy:        "admin",
			CreatedAt:        time.Now(),
		},
		{
			ID:               "lesson2",
			LanguageID:       "javascript",
			Title:            "Advanced Patterns",
			Difficulty:       2,
			Objectives:       []string{"Learn advanced JavaScript patterns"},
			Prerequisites:    []string{"lesson1"},
			EstimatedMinutes: 45,
			TokensCovered:    []string{"async", "await", "promise"},
			SnippetIDs:       []string{},
			Version:          1,
			CreatedBy:        "admin",
			CreatedAt:        time.Now(),
		},
	}

	for _, lesson := range lessons {
		m.lessons[lesson.ID] = lesson
	}

	// Add sample snippets
	snippets := []*models.Snippet{
		{
			ID:            "snippet1",
			LanguageID:    "javascript",
			Title:         "Hello World",
			SourceCode:    "console.log('Hello');",
			Tags:          []string{"test"},
			Difficulty:    1,
			EstimatedTime: 5,
			Checksum:      "abc123",
			AccessibilityTags: map[string]interface{}{
				"line_count": 10,
			},
			CreatedBy: "admin",
			Version:   1,
			CreatedAt: time.Now(),
		},
		{
			ID:            "snippet2",
			LanguageID:    "javascript",
			Title:         "Function Example",
			SourceCode:    "function test() {}",
			Tags:          []string{"test"},
			Difficulty:    1,
			EstimatedTime: 5,
			Checksum:      "def456",
			AccessibilityTags: map[string]interface{}{
				"line_count": 5,
			},
			CreatedBy: "admin",
			Version:   1,
			CreatedAt: time.Now(),
		},
	}

	for _, snippet := range snippets {
		m.snippets[snippet.ID] = snippet
	}
}

// Mock implementations for common operations
func (m *MockDatastore) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch v := src.(type) {
	case *models.User:
		if key == nil || key.Name == "" {
			key = datastore.NameKey("User", fmt.Sprintf("user_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.users[key.Name] = v
	case *models.Language:
		if key == nil {
			key = datastore.NameKey("Language", fmt.Sprintf("lang_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.languages[key.Name] = v
	case *models.Lesson:
		if key == nil || key.Name == "" {
			key = datastore.NameKey("Lesson", fmt.Sprintf("lesson_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.lessons[key.Name] = v
	case *models.Snippet:
		if key == nil || key.Name == "" {
			key = datastore.NameKey("Snippet", fmt.Sprintf("snippet_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.snippets[key.Name] = v
	case *models.Session:
		if key == nil || key.Name == "" {
			key = datastore.NameKey("Session", fmt.Sprintf("session_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.sessions[key.Name] = v
	case *models.Result:
		if key == nil {
			key = datastore.NameKey("Result", fmt.Sprintf("result_%d", time.Now().UnixNano()), nil)
		}
		v.SessionID = key.Name
		m.results[key.Name] = v
	}

	return key, nil
}

func (m *MockDatastore) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	if key == nil {
		return datastore.ErrNoSuchEntity
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	switch v := dst.(type) {
	case *models.User:
		if user, exists := m.users[key.Name]; exists {
			*v = *user
			return nil
		}
	case *models.Language:
		if lang, exists := m.languages[key.Name]; exists {
			*v = *lang
			return nil
		}
	case *models.Lesson:
		if lesson, exists := m.lessons[key.Name]; exists {
			*v = *lesson
			return nil
		}
	case *models.Snippet:
		if snippet, exists := m.snippets[key.Name]; exists {
			*v = *snippet
			return nil
		}
	case *models.Session:
		if session, exists := m.sessions[key.Name]; exists {
			*v = *session
			return nil
		}
	case *models.Result:
		if result, exists := m.results[key.Name]; exists {
			*v = *result
			return nil
		}
	}

	return datastore.ErrNoSuchEntity
}

func (m *MockDatastore) GetAll(ctx context.Context, q *datastore.Query, dst interface{}) ([]*datastore.Key, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// This is a simplified implementation that returns data based on the destination type
	var keys []*datastore.Key

	switch v := dst.(type) {
	case *[]models.User:
		users := v

		// For testing purposes, if this is a query for users, return all users
		// The Register handler will check for duplicates by email/handle in the returned slice
		for id, user := range m.users {
			*users = append(*users, *user)
			keys = append(keys, datastore.NameKey("User", id, nil))
		}
	case *[]*models.Language:
		languages := dst.(*[]*models.Language)
		for id, lang := range m.languages {
			*languages = append(*languages, lang)
			keys = append(keys, datastore.NameKey("Language", id, nil))
		}
	case *[]*models.Lesson:
		lessons := dst.(*[]*models.Lesson)
		for id, lesson := range m.lessons {
			*lessons = append(*lessons, lesson)
			keys = append(keys, datastore.NameKey("Lesson", id, nil))
		}
	case *[]*models.Snippet:
		snippets := dst.(*[]*models.Snippet)
		for id, snippet := range m.snippets {
			*snippets = append(*snippets, snippet)
			keys = append(keys, datastore.NameKey("Snippet", id, nil))
		}
	case *[]*models.Session:
		sessions := dst.(*[]*models.Session)
		for id, session := range m.sessions {
			*sessions = append(*sessions, session)
			keys = append(keys, datastore.NameKey("Session", id, nil))
		}
	}

	return keys, nil
}

func (m *MockDatastore) Count(ctx context.Context, q *datastore.Query) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// For simplicity, return counts for common entity types
	// In a real implementation, you'd need to inspect the query more carefully
	return len(m.users) + len(m.languages) + len(m.lessons) + len(m.snippets) + len(m.sessions), nil
}

func (m *MockDatastore) Delete(ctx context.Context, key *datastore.Key) error {
	if key == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	switch key.Kind {
	case "User":
		delete(m.users, key.Name)
	case "Language":
		delete(m.languages, key.Name)
	case "Lesson":
		delete(m.lessons, key.Name)
	case "Snippet":
		delete(m.snippets, key.Name)
	case "Session":
		delete(m.sessions, key.Name)
	case "Result":
		delete(m.results, key.Name)
	}

	return nil
}

func (m *MockDatastore) DeleteMulti(ctx context.Context, keys []*datastore.Key) error {
	if keys == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		if key == nil {
			continue
		}
		switch key.Kind {
		case "User":
			delete(m.users, key.Name)
		case "Language":
			delete(m.languages, key.Name)
		case "Lesson":
			delete(m.lessons, key.Name)
		case "Snippet":
			delete(m.snippets, key.Name)
		case "Session":
			delete(m.sessions, key.Name)
		case "SessionEvent":
			delete(m.events, key.Name)
		case "Result":
			delete(m.results, key.Name)
		case "LessonProgress":
			delete(m.lessonProgress, key.Name)
		case "Playlist":
			delete(m.playlists, key.Name)
		}
	}

	return nil
}

// Helper methods for testing

func (m *MockDatastore) PutUser(id string, user *models.User) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[id] = user
}

func (m *MockDatastore) PutLanguage(id string, language *models.Language) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.languages[id] = language
}

func (m *MockDatastore) PutLesson(id string, lesson *models.Lesson) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lessons[id] = lesson
}

func (m *MockDatastore) PutSnippet(id string, snippet *models.Snippet) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snippets[id] = snippet
}

func (m *MockDatastore) PutSession(id string, session *models.Session) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[id] = session
}

func (m *MockDatastore) PutPlaylist(id string, playlist *models.Playlist) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.playlists[id] = playlist
}

func (m *MockDatastore) PutLessonProgress(id string, progress *models.LessonProgress) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lessonProgress[id] = progress
}

func (m *MockDatastore) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users = make(map[string]*models.User)
	m.languages = make(map[string]*models.Language)
	m.lessons = make(map[string]*models.Lesson)
	m.snippets = make(map[string]*models.Snippet)
	m.sessions = make(map[string]*models.Session)
	m.events = make(map[string][]*models.SessionEvent)
	m.results = make(map[string]*models.Result)
	m.playlists = make(map[string]*models.Playlist)
	m.lessonProgress = make(map[string]*models.LessonProgress)
	m.contentVersions = make(map[string]*models.ContentVersion)
	m.contentValidations = make(map[string]*models.ContentValidation)
}
