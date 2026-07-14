package database

import (
	"context"
	"time"

	"github.com/typing-master-for-coding-backend/internal/models"
)

// InitializeDefaultAssessments creates the default assessment blueprints in the database
func InitializeDefaultAssessments(db *DatastoreClient) error {
	ctx := context.Background()

	assessments := []models.AssessmentBlueprint{
		{
			ID:                "javascript-basic-assessment",
			Name:              "JavaScript Basic Assessment",
			Description:       "Standardized assessment for JavaScript syntax and typing proficiency covering basic language constructs",
			Language:          "javascript",
			Difficulty:        3,
			EstimatedDuration: 15,
			SnippetIDs:        []string{"js-assessment-1", "js-assessment-2", "js-assessment-3"},
			PassingCriteria: models.AssessmentCriteria{
				MinimumAccuracy:          0.85,
				MinimumSpeed:             30,
				MaximumErrorRate:         5,
				StructuralAccuracyWeight: 0.8,
				SyntaxPenaltyMultiplier:  2.0,
				TimeLimit:                func() *int { t := 20; return &t }(),
			},
			ScoringWeights: models.AssessmentWeights{
				Speed:                0.25,
				Accuracy:             0.30,
				StructuralConformity: 0.20,
				SyntaxCorrectness:    0.15,
				Consistency:          0.05,
				ErrorRecovery:        0.05,
			},
			Version:   1,
			CreatedAt: time.Now(),
		},
		{
			ID:                "python-basic-assessment",
			Name:              "Python Basic Assessment",
			Description:       "Standardized assessment for Python syntax and typing proficiency covering basic language constructs and indentation",
			Language:          "python",
			Difficulty:        3,
			EstimatedDuration: 15,
			SnippetIDs:        []string{"py-assessment-1", "py-assessment-2", "py-assessment-3"},
			PassingCriteria: models.AssessmentCriteria{
				MinimumAccuracy:          0.85,
				MinimumSpeed:             28,
				MaximumErrorRate:         5,
				StructuralAccuracyWeight: 0.85, // Higher weight for Python due to indentation importance
				SyntaxPenaltyMultiplier:  2.5,
				TimeLimit:                func() *int { t := 20; return &t }(),
			},
			ScoringWeights: models.AssessmentWeights{
				Speed:                0.20,
				Accuracy:             0.30,
				StructuralConformity: 0.25, // Higher weight for Python
				SyntaxCorrectness:    0.15,
				Consistency:          0.05,
				ErrorRecovery:        0.05,
			},
			Version:   1,
			CreatedAt: time.Now(),
		},
		{
			ID:                "cpp-basic-assessment",
			Name:              "C++ Basic Assessment",
			Description:       "Standardized assessment for C++ syntax and typing proficiency covering basic language constructs and memory management",
			Language:          "cpp",
			Difficulty:        4,
			EstimatedDuration: 20,
			SnippetIDs:        []string{"cpp-assessment-1", "cpp-assessment-2", "cpp-assessment-3"},
			PassingCriteria: models.AssessmentCriteria{
				MinimumAccuracy:          0.80, // Lower due to complexity
				MinimumSpeed:             25,
				MaximumErrorRate:         6,
				StructuralAccuracyWeight: 0.75,
				SyntaxPenaltyMultiplier:  3.0, // Higher penalty for C++ syntax errors
				TimeLimit:                func() *int { t := 25; return &t }(),
			},
			ScoringWeights: models.AssessmentWeights{
				Speed:                0.20,
				Accuracy:             0.25,
				StructuralConformity: 0.25,
				SyntaxCorrectness:    0.20, // Higher weight due to complexity
				Consistency:          0.05,
				ErrorRecovery:        0.05,
			},
			Version:   1,
			CreatedAt: time.Now(),
		},
		{
			ID:                "rust-basic-assessment",
			Name:              "Rust Basic Assessment",
			Description:       "Standardized assessment for Rust syntax and typing proficiency covering ownership, borrowing, and basic language constructs",
			Language:          "rust",
			Difficulty:        4,
			EstimatedDuration: 20,
			SnippetIDs:        []string{"rust-assessment-1", "rust-assessment-2", "rust-assessment-3"},
			PassingCriteria: models.AssessmentCriteria{
				MinimumAccuracy:          0.80, // Lower due to complexity
				MinimumSpeed:             25,
				MaximumErrorRate:         6,
				StructuralAccuracyWeight: 0.75,
				SyntaxPenaltyMultiplier:  3.0, // Higher penalty for Rust syntax errors
				TimeLimit:                func() *int { t := 25; return &t }(),
			},
			ScoringWeights: models.AssessmentWeights{
				Speed:                0.20,
				Accuracy:             0.25,
				StructuralConformity: 0.25,
				SyntaxCorrectness:    0.20, // Higher weight due to complexity
				Consistency:          0.05,
				ErrorRecovery:        0.05,
			},
			Version:   1,
			CreatedAt: time.Now(),
		},
		{
			ID:                "yaml-basic-assessment",
			Name:              "YAML Basic Assessment",
			Description:       "Standardized assessment for YAML syntax and typing proficiency covering data structures and formatting",
			Language:          "yaml",
			Difficulty:        2,
			EstimatedDuration: 10,
			SnippetIDs:        []string{"yaml-assessment-1", "yaml-assessment-2", "yaml-assessment-3"},
			PassingCriteria: models.AssessmentCriteria{
				MinimumAccuracy:          0.90, // Higher due to simplicity
				MinimumSpeed:             35,
				MaximumErrorRate:         3,
				StructuralAccuracyWeight: 0.90, // Very high weight for YAML structure
				SyntaxPenaltyMultiplier:  2.0,
				TimeLimit:                func() *int { t := 15; return &t }(),
			},
			ScoringWeights: models.AssessmentWeights{
				Speed:                0.30,
				Accuracy:             0.35,
				StructuralConformity: 0.25,
				SyntaxCorrectness:    0.05,
				Consistency:          0.03,
				ErrorRecovery:        0.02,
			},
			Version:   1,
			CreatedAt: time.Now(),
		},
	}

	// Insert each assessment blueprint into the database
	for _, assessment := range assessments {
		key := NameKey("AssessmentBlueprint", assessment.ID, nil)

		// Check if assessment already exists
		var existing models.AssessmentBlueprint
		err := db.Get(ctx, key, &existing)
		if err == nil {
			// Assessment already exists, skip
			continue
		}

		// Create new assessment blueprint
		_, err = db.Put(ctx, key, &assessment)
		if err != nil {
			return err
		}
	}

	// Create default assessment snippets
	return initializeAssessmentSnippets(db)
}

// initializeAssessmentSnippets creates the default assessment code snippets
func initializeAssessmentSnippets(db *DatastoreClient) error {
	ctx := context.Background()

	snippets := []models.Snippet{
		// JavaScript Assessment Snippets
		{
			ID:         "js-assessment-1",
			LanguageID: "javascript",
			Title:      "JavaScript Functions and Variables",
			SourceCode: `function fibonacci(n) {
  if (n <= 1) return n;
  return fibonacci(n - 1) + fibonacci(n - 2);
}

const result = fibonacci(10);
console.log(result);`,
			Tags:              []string{"assessment", "functions", "recursion"},
			Difficulty:        3,
			EstimatedTime:     180,
			Checksum:          "js-assessment-1-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "camelCase": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "js-assessment-2",
			LanguageID: "javascript",
			Title:      "JavaScript Arrays and Objects",
			SourceCode: `const users = [
  { id: 1, name: 'Alice', active: true },
  { id: 2, name: 'Bob', active: false },
  { id: 3, name: 'Charlie', active: true }
];

const activeUsers = users
  .filter(user => user.active)
  .map(user => user.name);

console.log(activeUsers);`,
			Tags:              []string{"assessment", "arrays", "objects", "methods"},
			Difficulty:        3,
			EstimatedTime:     200,
			Checksum:          "js-assessment-2-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "camelCase": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "js-assessment-3",
			LanguageID: "javascript",
			Title:      "JavaScript Classes and Async",
			SourceCode: `class ApiClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }
  
  async fetchData(endpoint) {
    try {
      const response = await fetch(` + "`${this.baseUrl}/${endpoint}`" + `);
      return await response.json();
    } catch (error) {
      console.error('Fetch error:', error);
      throw error;
    }
  }
}`,
			Tags:              []string{"assessment", "classes", "async", "error-handling"},
			Difficulty:        4,
			EstimatedTime:     240,
			Checksum:          "js-assessment-3-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "camelCase": true},
			CreatedAt:         time.Now(),
		},

		// Python Assessment Snippets
		{
			ID:         "py-assessment-1",
			LanguageID: "python",
			Title:      "Python Functions and List Comprehensions",
			SourceCode: `def process_numbers(numbers):
    """Process a list of numbers and return even squares."""
    return [x**2 for x in numbers if x % 2 == 0]

def main():
    data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
    result = process_numbers(data)
    print(f"Even squares: {result}")

if __name__ == "__main__":
    main()`,
			Tags:              []string{"assessment", "functions", "list-comprehensions"},
			Difficulty:        3,
			EstimatedTime:     180,
			Checksum:          "py-assessment-1-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "indentation": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "py-assessment-2",
			LanguageID: "python",
			Title:      "Python Classes and Exception Handling",
			SourceCode: `class Calculator:
    def __init__(self):
        self.history = []
    
    def divide(self, a, b):
        try:
            result = a / b
            self.history.append(f"{a} / {b} = {result}")
            return result
        except ZeroDivisionError:
            print("Error: Division by zero")
            return None
    
    def get_history(self):
        return self.history.copy()`,
			Tags:              []string{"assessment", "classes", "exceptions"},
			Difficulty:        3,
			EstimatedTime:     200,
			Checksum:          "py-assessment-2-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "indentation": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "py-assessment-3",
			LanguageID: "python",
			Title:      "Python Decorators and Context Managers",
			SourceCode: `import time
from contextlib import contextmanager

def timing_decorator(func):
    def wrapper(*args, **kwargs):
        start = time.time()
        result = func(*args, **kwargs)
        end = time.time()
        print(f"{func.__name__} took {end - start:.4f} seconds")
        return result
    return wrapper

@contextmanager
def database_connection():
    print("Opening database connection")
    try:
        yield "connection"
    finally:
        print("Closing database connection")`,
			Tags:              []string{"assessment", "decorators", "context-managers"},
			Difficulty:        4,
			EstimatedTime:     240,
			Checksum:          "py-assessment-3-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "indentation": true},
			CreatedAt:         time.Now(),
		},

		// C++ Assessment Snippets
		{
			ID:         "cpp-assessment-1",
			LanguageID: "cpp",
			Title:      "C++ Classes and Memory Management",
			SourceCode: `#include <iostream>
#include <memory>

class Vector {
private:
    std::unique_ptr<double[]> data;
    size_t size;

public:
    Vector(size_t n) : size(n), data(std::make_unique<double[]>(n)) {}
    
    double& operator[](size_t index) {
        return data[index];
    }
    
    size_t getSize() const {
        return size;
    }
};`,
			Tags:              []string{"assessment", "classes", "smart-pointers"},
			Difficulty:        4,
			EstimatedTime:     220,
			Checksum:          "cpp-assessment-1-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "pointers": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "cpp-assessment-2",
			LanguageID: "cpp",
			Title:      "C++ Templates and STL",
			SourceCode: `#include <vector>
#include <algorithm>
#include <iostream>

template<typename T>
class Stack {
private:
    std::vector<T> elements;

public:
    void push(const T& element) {
        elements.push_back(element);
    }
    
    T pop() {
        if (elements.empty()) {
            throw std::runtime_error("Stack is empty");
        }
        T top = elements.back();
        elements.pop_back();
        return top;
    }
    
    bool empty() const {
        return elements.empty();
    }
};`,
			Tags:              []string{"assessment", "templates", "stl", "exceptions"},
			Difficulty:        4,
			EstimatedTime:     240,
			Checksum:          "cpp-assessment-2-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "templates": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "cpp-assessment-3",
			LanguageID: "cpp",
			Title:      "C++ Move Semantics and RAII",
			SourceCode: `#include <utility>
#include <iostream>

class Resource {
private:
    int* data;
    size_t size;

public:
    Resource(size_t n) : size(n), data(new int[n]) {
        std::cout << "Resource acquired\n";
    }
    
    ~Resource() {
        delete[] data;
        std::cout << "Resource released\n";
    }
    
    Resource(Resource&& other) noexcept 
        : data(std::exchange(other.data, nullptr)), 
          size(std::exchange(other.size, 0)) {}
    
    Resource& operator=(Resource&& other) noexcept {
        if (this != &other) {
            delete[] data;
            data = std::exchange(other.data, nullptr);
            size = std::exchange(other.size, 0);
        }
        return *this;
    }
};`,
			Tags:              []string{"assessment", "move-semantics", "raii"},
			Difficulty:        5,
			EstimatedTime:     280,
			Checksum:          "cpp-assessment-3-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "very-high", "move-semantics": true},
			CreatedAt:         time.Now(),
		},

		// Rust Assessment Snippets
		{
			ID:         "rust-assessment-1",
			LanguageID: "rust",
			Title:      "Rust Ownership and Borrowing",
			SourceCode: `fn main() {
    let mut numbers = vec![1, 2, 3, 4, 5];
    
    let sum = calculate_sum(&numbers);
    println!("Sum: {}", sum);
    
    modify_vector(&mut numbers);
    println!("Modified: {:?}", numbers);
}

fn calculate_sum(nums: &[i32]) -> i32 {
    nums.iter().sum()
}

fn modify_vector(nums: &mut Vec<i32>) {
    nums.push(6);
    nums.iter_mut().for_each(|x| *x *= 2);
}`,
			Tags:              []string{"assessment", "ownership", "borrowing"},
			Difficulty:        4,
			EstimatedTime:     200,
			Checksum:          "rust-assessment-1-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "ownership": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "rust-assessment-2",
			LanguageID: "rust",
			Title:      "Rust Pattern Matching and Error Handling",
			SourceCode: `use std::fs::File;
use std::io::Read;

#[derive(Debug)]
enum ParseError {
    InvalidFormat,
    MissingData,
}

fn parse_config(filename: &str) -> Result<Config, ParseError> {
    let mut file = File::open(filename)
        .map_err(|_| ParseError::MissingData)?;
    
    let mut contents = String::new();
    file.read_to_string(&mut contents)
        .map_err(|_| ParseError::InvalidFormat)?;
    
    match contents.trim() {
        "debug" => Ok(Config::Debug),
        "release" => Ok(Config::Release),
        _ => Err(ParseError::InvalidFormat),
    }
}

#[derive(Debug)]
enum Config {
    Debug,
    Release,
}`,
			Tags:              []string{"assessment", "pattern-matching", "error-handling"},
			Difficulty:        4,
			EstimatedTime:     240,
			Checksum:          "rust-assessment-2-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "pattern-matching": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "rust-assessment-3",
			LanguageID: "rust",
			Title:      "Rust Traits and Generics",
			SourceCode: `trait Drawable {
    fn draw(&self);
    fn area(&self) -> f64;
}

struct Circle {
    radius: f64,
}

struct Rectangle {
    width: f64,
    height: f64,
}

impl Drawable for Circle {
    fn draw(&self) {
        println!("Drawing circle with radius {}", self.radius);
    }
    
    fn area(&self) -> f64 {
        std::f64::consts::PI * self.radius * self.radius
    }
}

impl Drawable for Rectangle {
    fn draw(&self) {
        println!("Drawing rectangle {}x{}", self.width, self.height);
    }
    
    fn area(&self) -> f64 {
        self.width * self.height
    }
}

fn print_info<T: Drawable>(shape: &T) {
    shape.draw();
    println!("Area: {:.2}", shape.area());
}`,
			Tags:              []string{"assessment", "traits", "generics"},
			Difficulty:        4,
			EstimatedTime:     260,
			Checksum:          "rust-assessment-3-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "high", "traits": true},
			CreatedAt:         time.Now(),
		},

		// YAML Assessment Snippets
		{
			ID:         "yaml-assessment-1",
			LanguageID: "yaml",
			Title:      "YAML Basic Configuration",
			SourceCode: `# Application Configuration
app:
  name: "typing-master"
  version: "1.0.0"
  debug: true

database:
  host: "localhost"
  port: 5432
  name: "typing_db"
  credentials:
    username: "admin"
    password: "secret123"

features:
  - authentication
  - leaderboards
  - assessments
  - offline_mode`,
			Tags:              []string{"assessment", "configuration", "basic-structure"},
			Difficulty:        2,
			EstimatedTime:     120,
			Checksum:          "yaml-assessment-1-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "low", "indentation": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "yaml-assessment-2",
			LanguageID: "yaml",
			Title:      "YAML Docker Compose",
			SourceCode: `version: '3.8'

services:
  web:
    build: .
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=production
      - DATABASE_URL=postgresql://user:pass@db:5432/app
    depends_on:
      - db
      - redis

  db:
    image: postgres:13
    environment:
      POSTGRES_DB: app
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:`,
			Tags:              []string{"assessment", "docker-compose", "services"},
			Difficulty:        3,
			EstimatedTime:     180,
			Checksum:          "yaml-assessment-2-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "indentation": true},
			CreatedAt:         time.Now(),
		},
		{
			ID:         "yaml-assessment-3",
			LanguageID: "yaml",
			Title:      "YAML Kubernetes Deployment",
			SourceCode: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: typing-master-app
  labels:
    app: typing-master
spec:
  replicas: 3
  selector:
    matchLabels:
      app: typing-master
  template:
    metadata:
      labels:
        app: typing-master
    spec:
      containers:
      - name: app
        image: typing-master:latest
        ports:
        - containerPort: 3000
        env:
        - name: NODE_ENV
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: db-secret
              key: url
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"`,
			Tags:              []string{"assessment", "kubernetes", "deployment"},
			Difficulty:        3,
			EstimatedTime:     200,
			Checksum:          "yaml-assessment-3-checksum",
			AccessibilityTags: map[string]interface{}{"complexity": "medium", "indentation": true},
			CreatedAt:         time.Now(),
		},
	}

	// Insert each snippet into the database
	for _, snippet := range snippets {
		key := NameKey("Snippet", snippet.ID, nil)

		// Check if snippet already exists
		var existing models.Snippet
		err := db.Get(ctx, key, &existing)
		if err == nil {
			// Snippet already exists, skip
			continue
		}

		// Create new snippet
		_, err = db.Put(ctx, key, &snippet)
		if err != nil {
			return err
		}
	}

	return nil
}
