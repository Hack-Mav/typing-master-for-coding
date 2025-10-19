package database

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"typing-master-backend/internal/models"
)

// InitializeDefaultLanguages creates the default language configurations in Datastore
func InitializeDefaultLanguages(db *DatastoreClient) error {
	ctx := context.Background()

	languages := []models.Language{
		{
			ID:       "python",
			Name:     "Python",
			Version:  1,
			ParserID: "tree-sitter-python",
			GrammarConfig: map[string]interface{}{
				"tokenTypes": []string{"identifier", "string", "number", "keyword", "operator", "delimiter", "comment"},
				"keywords":   []string{"def", "class", "if", "else", "elif", "for", "while", "try", "except", "import", "from", "return"},
				"operators":  []string{"+", "-", "*", "/", "//", "%", "**", "==", "!=", "<", ">", "<=", ">=", "and", "or", "not"},
				"delimiters": []string{"(", ")", "[", "]", "{", "}", ":", ",", "."},
			},
			WhitespaceRules: map[string]interface{}{
				"indentSize":             4,
				"useSpaces":              true,
				"trimTrailingWhitespace": true,
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
		},
		{
			ID:       "javascript",
			Name:     "JavaScript",
			Version:  1,
			ParserID: "tree-sitter-javascript",
			GrammarConfig: map[string]interface{}{
				"tokenTypes": []string{"identifier", "string", "number", "keyword", "operator", "delimiter", "comment"},
				"keywords":   []string{"function", "const", "let", "var", "if", "else", "for", "while", "return", "class", "import", "export"},
				"operators":  []string{"+", "-", "*", "/", "%", "==", "===", "!=", "!==", "<", ">", "<=", ">=", "&&", "||", "!"},
				"delimiters": []string{"(", ")", "[", "]", "{", "}", ";", ",", "."},
			},
			WhitespaceRules: map[string]interface{}{
				"indentSize":             2,
				"useSpaces":              true,
				"trimTrailingWhitespace": true,
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
		},
		{
			ID:       "yaml",
			Name:     "YAML",
			Version:  1,
			ParserID: "tree-sitter-yaml",
			GrammarConfig: map[string]interface{}{
				"tokenTypes": []string{"key", "value", "string", "number", "boolean", "null", "comment"},
				"keywords":   []string{"true", "false", "null"},
				"operators":  []string{":", "-", "|", ">"},
				"delimiters": []string{"[", "]", "{", "}", ","},
			},
			WhitespaceRules: map[string]interface{}{
				"indentSize":             2,
				"useSpaces":              true,
				"trimTrailingWhitespace": true,
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
		},
		{
			ID:       "cpp",
			Name:     "C++",
			Version:  1,
			ParserID: "tree-sitter-cpp",
			GrammarConfig: map[string]interface{}{
				"tokenTypes": []string{"identifier", "string", "number", "keyword", "operator", "delimiter", "comment", "preprocessor", "type"},
				"keywords": []string{
					"auto", "bool", "break", "case", "catch", "char", "class", "const", "continue", "default",
					"delete", "do", "double", "else", "enum", "explicit", "extern", "false", "float", "for",
					"friend", "goto", "if", "inline", "int", "long", "namespace", "new", "nullptr", "operator",
					"private", "protected", "public", "return", "short", "signed", "sizeof", "static", "struct",
					"switch", "template", "this", "throw", "true", "try", "typedef", "typename", "union",
					"unsigned", "using", "virtual", "void", "volatile", "while",
				},
				"operators": []string{
					"+", "-", "*", "/", "%", "++", "--", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "!",
					"&", "|", "^", "~", "<<", ">>", "=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=",
					"<<=", ">>=", "->", "::", ".*", "->*",
				},
				"delimiters": []string{"(", ")", "[", "]", "{", "}", ";", ",", ".", "<", ">"},
			},
			WhitespaceRules: map[string]interface{}{
				"indentSize":             4,
				"useSpaces":              true,
				"trimTrailingWhitespace": true,
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
		},
		{
			ID:       "rust",
			Name:     "Rust",
			Version:  1,
			ParserID: "tree-sitter-rust",
			GrammarConfig: map[string]interface{}{
				"tokenTypes": []string{"identifier", "string", "number", "keyword", "operator", "delimiter", "comment", "attribute", "lifetime", "type"},
				"keywords": []string{
					"as", "async", "await", "break", "const", "continue", "crate", "dyn", "else", "enum",
					"extern", "false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod",
					"move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct", "super",
					"trait", "true", "type", "unsafe", "use", "where", "while", "abstract", "become",
					"box", "do", "final", "macro", "override", "priv", "typeof", "unsized", "virtual", "yield",
				},
				"operators": []string{
					"+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "!", "&", "|",
					"^", "<<", ">>", "=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>=",
					"->", "=>", "::", "..", "..=", "?",
				},
				"delimiters": []string{"(", ")", "[", "]", "{", "}", ";", ",", ".", "<", ">"},
			},
			WhitespaceRules: map[string]interface{}{
				"indentSize":             4,
				"useSpaces":              true,
				"trimTrailingWhitespace": true,
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
		},
	}

	// Insert each language into Datastore
	for _, language := range languages {
		key := datastore.NameKey("Language", language.ID, nil)

		// Check if language already exists
		var existing models.Language
		err := db.Get(ctx, key, &existing)
		if err == nil {
			// Language already exists, skip
			continue
		}

		// Create new language
		_, err = db.Put(ctx, key, &language)
		if err != nil {
			return err
		}
	}

	return nil
}