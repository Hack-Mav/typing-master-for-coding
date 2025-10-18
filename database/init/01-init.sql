-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create initial languages
INSERT INTO languages (id, name, version, parser_id, grammar_config, whitespace_rules) VALUES
('python', 'Python', '3.11', 'tree-sitter-python', '{"indent_size": 4, "use_tabs": false}', '{"enforce_pep8": true, "max_line_length": 88}'),
('javascript', 'JavaScript', 'ES2023', 'tree-sitter-javascript', '{"indent_size": 2, "use_tabs": false}', '{"semicolons": "optional", "trailing_commas": true}'),
('yaml', 'YAML', '1.2', 'tree-sitter-yaml', '{"indent_size": 2, "use_tabs": false}', '{"enforce_indentation": true}'),
('cpp', 'C++', '20', 'tree-sitter-cpp', '{"indent_size": 2, "use_tabs": false}', '{"brace_style": "allman"}'),
('rust', 'Rust', '1.70', 'tree-sitter-rust', '{"indent_size": 4, "use_tabs": false}', '{"enforce_rustfmt": true}')
ON CONFLICT (id) DO NOTHING;

-- Create sample snippets for each language
INSERT INTO snippets (language_id, title, source_code, tags, difficulty, estimated_time, accessibility_tags) VALUES
('python', 'Hello World', 'print("Hello, World!")', ARRAY['basic', 'intro'], 1, 30, '{"complexity": "low", "camelCase": false}'),
('javascript', 'Arrow Function', 'const greet = (name) => `Hello, ${name}!`;', ARRAY['functions', 'es6'], 2, 45, '{"complexity": "medium", "camelCase": true}'),
('yaml', 'Basic Config', 'name: typing-master\nversion: 1.0.0\nports:\n  - 3000\n  - 8080', ARRAY['config', 'basic'], 1, 60, '{"complexity": "low", "indentation": "strict"}'),
('cpp', 'Simple Class', 'class Rectangle {\npublic:\n    int width, height;\n    int area() { return width * height; }\n};', ARRAY['oop', 'classes'], 3, 90, '{"complexity": "high", "punctuation": "heavy"}'),
('rust', 'Struct Definition', 'struct Point {\n    x: f64,\n    y: f64,\n}\n\nimpl Point {\n    fn new(x: f64, y: f64) -> Point {\n        Point { x, y }\n    }\n}', ARRAY['structs', 'impl'], 3, 120, '{"complexity": "high", "ownership": true}')
ON CONFLICT (id) DO NOTHING;