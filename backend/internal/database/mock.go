package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"cloud.google.com/go/datastore"
	"typing-master-backend/internal/models"
)

// MockDatastore provides an in-memory implementation for development
type MockDatastore struct {
	mu        sync.RWMutex
	users     map[string]*models.User
	languages map[string]*models.Language
	lessons   map[string]*models.Lesson
	snippets  map[string]*models.Snippet
	sessions  map[string]*models.Session
	events    map[string][]*models.SessionEvent
	results   map[string]*models.Result
}

func NewMockDatastore() *MockDatastore {
	mock := &MockDatastore{
		users:     make(map[string]*models.User),
		languages: make(map[string]*models.Language),
		lessons:   make(map[string]*models.Lesson),
		snippets:  make(map[string]*models.Snippet),
		sessions:  make(map[string]*models.Session),
		events:    make(map[string][]*models.SessionEvent),
		results:   make(map[string]*models.Result),
	}
	
	// Initialize with some sample data
	mock.initSampleData()
	
	return mock
}

func (m *MockDatastore) initSampleData() {
	// Add sample languages
	languages := []*models.Language{
		{
			ID:              "javascript",
			Name:            "JavaScript",
			Version:         "ES2023",
			ParserID:        "tree-sitter-javascript",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedAt:       time.Now(),
		},
		{
			ID:              "python",
			Name:            "Python",
			Version:         "3.11",
			ParserID:        "tree-sitter-python",
			GrammarConfig:   map[string]interface{}{"indentation": "spaces"},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedAt:       time.Now(),
		},
		{
			ID:              "cpp",
			Name:            "C++",
			Version:         "C++20",
			ParserID:        "tree-sitter-cpp",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedAt:       time.Now(),
		},
	}
	
	for _, lang := range languages {
		m.languages[lang.ID] = lang
	}
}

// Mock implementations for common operations
func (m *MockDatastore) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	switch v := src.(type) {
	case *models.User:
		if key.Name == "" {
			key = datastore.NameKey("User", fmt.Sprintf("user_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.users[key.Name] = v
	case *models.Language:
		v.ID = key.Name
		m.languages[key.Name] = v
	case *models.Lesson:
		if key.Name == "" {
			key = datastore.NameKey("Lesson", fmt.Sprintf("lesson_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.lessons[key.Name] = v
	case *models.Snippet:
		if key.Name == "" {
			key = datastore.NameKey("Snippet", fmt.Sprintf("snippet_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.snippets[key.Name] = v
	case *models.Session:
		if key.Name == "" {
			key = datastore.NameKey("Session", fmt.Sprintf("session_%d", time.Now().UnixNano()), nil)
		}
		v.ID = key.Name
		m.sessions[key.Name] = v
	case *models.Result:
		v.SessionID = key.Name
		m.results[key.Name] = v
	}
	
	return key, nil
}

func (m *MockDatastore) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
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
	
	// This is a simplified implementation that returns all data based on the destination type
	var keys []*datastore.Key
	
	switch dst.(type) {
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

func (m *MockDatastore) Delete(ctx context.Context, key *datastore.Key) error {
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