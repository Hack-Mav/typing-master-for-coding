package database

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/typing-master-for-coding-backend/internal/models"
)

// MockDatastore is an in-memory datastore used for tests and local development.
type MockDatastore struct {
	mu       sync.RWMutex
	entities map[string]map[string]json.RawMessage
}

// NewMockDatastore creates a new in-memory datastore with sample data.
func NewMockDatastore() *MockDatastore {
	m := &MockDatastore{
		entities: make(map[string]map[string]json.RawMessage),
	}
	m.initSampleData()
	return m
}

func (m *MockDatastore) initSampleData() {
	languages := []models.Language{
		{
			ID:              "javascript",
			Name:            "JavaScript",
			Version:         "1",
			ParserID:        "tree-sitter-javascript",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "python",
			Name:            "Python",
			Version:         "1",
			ParserID:        "tree-sitter-python",
			GrammarConfig:   map[string]interface{}{"indentation": "spaces"},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "cpp",
			Name:            "C++",
			Version:         "1",
			ParserID:        "tree-sitter-cpp",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "rust",
			Name:            "Rust",
			Version:         "1",
			ParserID:        "tree-sitter-rust",
			GrammarConfig:   map[string]interface{}{"semicolons": true},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
		{
			ID:              "yaml",
			Name:            "YAML",
			Version:         "1",
			ParserID:        "tree-sitter-yaml",
			GrammarConfig:   map[string]interface{}{"indentation": "spaces"},
			WhitespaceRules: map[string]interface{}{"indentation": "spaces"},
			CreatedBy:       "admin",
			CreatedAt:       time.Now(),
		},
	}

	for _, lang := range languages {
		key := NameKey("Language", lang.ID, nil)
		_, _ = m.Put(context.Background(), key, &lang)
	}

	snippets := []models.Snippet{
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
		key := NameKey("Snippet", snippet.ID, nil)
		_, _ = m.Put(context.Background(), key, &snippet)
	}
}

// Put stores an entity with the provided key.
func (m *MockDatastore) Put(ctx context.Context, key *Key, src interface{}) (*Key, error) {
	if key == nil {
		return nil, ErrNoSuchEntity
	}
	if src == nil {
		return nil, fmt.Errorf("database: src is nil")
	}

	data, err := marshalEntity(src)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.entities[key.Kind] == nil {
		m.entities[key.Kind] = make(map[string]json.RawMessage)
	}
	m.entities[key.Kind][key.Name] = data
	return key, nil
}

// PutMulti stores multiple entities.
func (m *MockDatastore) PutMulti(ctx context.Context, keys []*Key, src interface{}) ([]*Key, error) {
	v := reflect.ValueOf(src)
	if v.Kind() != reflect.Slice {
		return nil, fmt.Errorf("database: src must be a slice")
	}
	if len(keys) != v.Len() {
		return nil, fmt.Errorf("database: keys and src length mismatch")
	}

	returnedKeys := make([]*Key, len(keys))
	for i, key := range keys {
		elem := v.Index(i).Interface()
		newKey, err := m.Put(ctx, key, elem)
		if err != nil {
			return nil, err
		}
		returnedKeys[i] = newKey
	}
	return returnedKeys, nil
}

// Get retrieves an entity by key.
func (m *MockDatastore) Get(ctx context.Context, key *Key, dst interface{}) error {
	if key == nil {
		return ErrNoSuchEntity
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	data, exists := m.entities[key.Kind][key.Name]
	if !exists {
		return ErrNoSuchEntity
	}

	return unmarshalEntity(data, dst)
}

// GetAll retrieves entities matching the query.
func (m *MockDatastore) GetAll(ctx context.Context, query *Query, dst interface{}) ([]*Key, error) {
	if query.err != nil {
		return nil, query.err
	}

	keys, data, err := m.query(ctx, query)
	if err != nil {
		return nil, err
	}

	if query.keysOnly || dst == nil {
		return keys, nil
	}

	if err := unmarshalSlice(dst, data); err != nil {
		return nil, err
	}
	return keys, nil
}

// Run returns an iterator over the query results.
func (m *MockDatastore) Run(ctx context.Context, query *Query) Iterator {
	keys, data, err := m.query(ctx, query)
	if err != nil {
		return &errorIterator{err: err}
	}
	return newSliceIterator(keys, data)
}

// Delete removes an entity by key.
func (m *MockDatastore) Delete(ctx context.Context, key *Key) error {
	if key == nil {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.entities[key.Kind] != nil {
		delete(m.entities[key.Kind], key.Name)
	}
	return nil
}

// DeleteMulti removes multiple entities by key.
func (m *MockDatastore) DeleteMulti(ctx context.Context, keys []*Key) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

// Count returns the number of entities matching the query.
func (m *MockDatastore) Count(ctx context.Context, query *Query) (int, error) {
	if query.err != nil {
		return 0, query.err
	}

	keys, _, err := m.query(ctx, query)
	if err != nil {
		return 0, err
	}
	return len(keys), nil
}

func (m *MockDatastore) query(ctx context.Context, query *Query) ([]*Key, [][]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := m.entities[query.kind]
	if items == nil {
		return nil, nil, nil
	}

	var keys []*Key
	var data [][]byte
	for name, payload := range items {
		keys = append(keys, &Key{Kind: query.kind, Name: name})
		data = append(data, payload)
	}

	// Apply simple in-memory filters (currently best-effort; keeps tests stable).
	if len(query.filters) > 0 {
		filteredKeys := make([]*Key, 0, len(keys))
		filteredData := make([][]byte, 0, len(keys))
		for i, payload := range data {
			if m.matchesFilters(payload, query.filters) {
				filteredKeys = append(filteredKeys, keys[i])
				filteredData = append(filteredData, payload)
			}
		}
		keys = filteredKeys
		data = filteredData
	}

	// Apply order (best-effort; not critical for tests).
	if len(query.orders) > 0 {
		m.sortResults(data, keys, query.orders)
	}

	// Apply offset and limit.
	if query.offset > 0 && query.offset < len(keys) {
		keys = keys[query.offset:]
		data = data[query.offset:]
	} else if query.offset >= len(keys) {
		return nil, nil, nil
	}

	if query.limit > 0 && query.limit < len(keys) {
		keys = keys[:query.limit]
		data = data[:query.limit]
	}

	return keys, data, nil
}

func (m *MockDatastore) matchesFilters(payload json.RawMessage, filters []filter) bool {
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return false
	}

	for _, f := range filters {
		key := toSnakeCase(f.fieldName)
		val, ok := raw[key]
		if !ok {
			return false
		}
		if !compareFilterValue(val, f.value, f.operator) {
			return false
		}
	}
	return true
}

func compareFilterValue(dataValue, filterValue interface{}, op string) bool {
	op = normalizeOpString(op)

	ds := fmt.Sprintf("%v", dataValue)
	fs := fmt.Sprintf("%v", filterValue)

	switch op {
	case "=":
		return ds == fs
	case "!=":
		return ds != fs
	case ">", ">=", "<", "<=":
		return compareComparable(dataValue, filterValue, op)
	default:
		return false
	}
}

func normalizeOpString(op string) string {
	switch op {
	case "=", "==":
		return "="
	case ">":
		return ">"
	case ">=":
		return ">="
	case "<":
		return "<"
	case "<=":
		return "<="
	case "!=", "<>":
		return "!="
	}
	return "="
}

func compareComparable(dataValue, filterValue interface{}, op string) bool {
	// Try time comparison first.
	if dt, ok := parseTime(dataValue); ok {
		if ft, ok := parseTime(filterValue); ok {
			switch op {
			case ">":
				return dt.After(ft)
			case ">=":
				return dt.After(ft) || dt.Equal(ft)
			case "<":
				return dt.Before(ft)
			case "<=":
				return dt.Before(ft) || dt.Equal(ft)
			}
		}
	}

	// Try numeric comparison.
	if dn, ok := parseNumber(dataValue); ok {
		if fn, ok := parseNumber(filterValue); ok {
			switch op {
			case ">":
				return dn > fn
			case ">=":
				return dn >= fn
			case "<":
				return dn < fn
			case "<=":
				return dn <= fn
			}
		}
	}

	// Fall back to string comparison.
	ds := fmt.Sprintf("%v", dataValue)
	fs := fmt.Sprintf("%v", filterValue)
	switch op {
	case ">":
		return ds > fs
	case ">=":
		return ds >= fs
	case "<":
		return ds < fs
	case "<=":
		return ds <= fs
	}
	return false
}

func parseTime(v interface{}) (time.Time, bool) {
	switch t := v.(type) {
	case time.Time:
		return t, true
	case string:
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z", "2006-01-02T15:04:05.999Z"} {
			if parsed, err := time.Parse(layout, t); err == nil {
				return parsed, true
			}
		}
	}
	return time.Time{}, false
}

func parseNumber(v interface{}) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	case string:
		if parsed, err := parseNumberString(n); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

func parseNumberString(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

func (m *MockDatastore) sortResults(data [][]byte, keys []*Key, orders []order) {
	if len(orders) == 0 {
		return
	}
	raw := make([]map[string]interface{}, len(data))
	for i, payload := range data {
		var obj map[string]interface{}
		_ = json.Unmarshal(payload, &obj)
		raw[i] = obj
	}

	sort.Slice(keys, func(i, j int) bool {
		for _, o := range orders {
			key := toSnakeCase(o.fieldName)
			vi, vj := raw[i][key], raw[j][key]
			cmp := compareValues(vi, vj)
			if cmp != 0 {
				if o.descending {
					return cmp > 0
				}
				return cmp < 0
			}
		}
		return false
	})

	// Reorder data alongside keys.
	newData := make([][]byte, len(data))
	for i, key := range keys {
		for j, k := range keys {
			if k == key {
				newData[i] = data[j]
				break
			}
		}
	}
	copy(data, newData)
}

func compareValues(a, b interface{}) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	if an, ok := parseNumber(a); ok {
		if bn, ok := parseNumber(b); ok {
			if an < bn {
				return -1
			}
			if an > bn {
				return 1
			}
			return 0
		}
	}

	if at, ok := parseTime(a); ok {
		if bt, ok := parseTime(b); ok {
			if at.Before(bt) {
				return -1
			}
			if at.After(bt) {
				return 1
			}
			return 0
		}
	}

	sa := fmt.Sprintf("%v", a)
	sb := fmt.Sprintf("%v", b)
	if sa < sb {
		return -1
	}
	if sa > sb {
		return 1
	}
	return 0
}

// Helper methods for testing

func (m *MockDatastore) PutUser(id string, user *models.User) {
	user.ID = id
	_, _ = m.Put(context.Background(), NameKey("User", id, nil), user)
}

func (m *MockDatastore) PutLanguage(id string, language *models.Language) {
	language.ID = id
	_, _ = m.Put(context.Background(), NameKey("Language", id, nil), language)
}

func (m *MockDatastore) PutLesson(id string, lesson *models.Lesson) {
	lesson.ID = id
	_, _ = m.Put(context.Background(), NameKey("Lesson", id, nil), lesson)
}

func (m *MockDatastore) PutSnippet(id string, snippet *models.Snippet) {
	snippet.ID = id
	_, _ = m.Put(context.Background(), NameKey("Snippet", id, nil), snippet)
}

func (m *MockDatastore) PutSession(id string, session *models.Session) {
	session.ID = id
	_, _ = m.Put(context.Background(), NameKey("Session", id, nil), session)
}

func (m *MockDatastore) PutPlaylist(id string, playlist *models.Playlist) {
	playlist.ID = id
	_, _ = m.Put(context.Background(), NameKey("Playlist", id, nil), playlist)
}

func (m *MockDatastore) PutLessonProgress(id string, progress *models.LessonProgress) {
	progress.ID = id
	_, _ = m.Put(context.Background(), NameKey("LessonProgress", id, nil), progress)
}

// Tutorial operations
func (m *MockDatastore) GetAllTutorials() ([]models.Tutorial, error) {
	var tutorials []models.Tutorial
	_, err := m.GetAll(context.Background(), NewQuery("Tutorial"), &tutorials)
	if err != nil {
		return nil, err
	}

	var active []models.Tutorial
	for _, t := range tutorials {
		if t.IsActive {
			active = append(active, t)
		}
	}
	return active, nil
}

func (m *MockDatastore) GetTutorialsByCategory(category string) ([]models.Tutorial, error) {
	var tutorials []models.Tutorial
	_, err := m.GetAll(context.Background(), NewQuery("Tutorial").FilterField("Category", "=", category), &tutorials)
	if err != nil {
		return nil, err
	}

	var filtered []models.Tutorial
	for _, t := range tutorials {
		if t.IsActive && t.Category == category {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *MockDatastore) GetTutorialsByDifficulty(difficulty string) ([]models.Tutorial, error) {
	var tutorials []models.Tutorial
	_, err := m.GetAll(context.Background(), NewQuery("Tutorial").FilterField("Difficulty", "=", difficulty), &tutorials)
	if err != nil {
		return nil, err
	}

	var filtered []models.Tutorial
	for _, t := range tutorials {
		if t.IsActive && t.Difficulty == difficulty {
			filtered = append(filtered, t)
		}
	}
	return filtered, nil
}

func (m *MockDatastore) GetTutorialByID(tutorialID string) (*models.Tutorial, error) {
	var tutorial models.Tutorial
	if err := m.Get(context.Background(), NameKey("Tutorial", tutorialID, nil), &tutorial); err != nil {
		return nil, err
	}
	if !tutorial.IsActive {
		return nil, ErrNoSuchEntity
	}
	return &tutorial, nil
}

// Clear removes all data from the mock datastore.
func (m *MockDatastore) Clear() {
	m.mu.Lock()
	m.entities = make(map[string]map[string]json.RawMessage)
	m.mu.Unlock()
	m.initSampleData()
}

// Ensure MockDatastore implements storage.
var _ storage = (*MockDatastore)(nil)
