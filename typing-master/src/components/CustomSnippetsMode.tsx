import React, { useState, useCallback, useEffect } from 'react';
import MonacoTypingInterface from './MonacoTypingInterface';
import { SessionResult } from '../types/session';
import { Language } from '../types/parser';
import { contentService } from '../services/ContentService';
import { authService } from '../services/AuthService';
import './CustomSnippetsMode.css';

interface CustomSnippetsModeProps {
  languageId: Language;
  onComplete?: (result: SessionResult) => void;
  onExit?: () => void;
}

interface CustomSnippet {
  id: string;
  code: string;
  language: Language;
  name: string;
  createdAt: Date;
  userId?: string;
  tags?: string[];
  difficulty?: number;
}

const CustomSnippetsMode: React.FC<CustomSnippetsModeProps> = ({
  languageId,
  onComplete,
  onExit,
}) => {
  const [showEditor, setShowEditor] = useState(false);
  const [customCode, setCustomCode] = useState('');
  const [snippetName, setSnippetName] = useState('');
  const [selectedLanguage, setSelectedLanguage] =
    useState<Language>(languageId);
  const [savedSnippets, setSavedSnippets] = useState<CustomSnippet[]>([]);
  const [activeSnippet, setActiveSnippet] = useState<CustomSnippet | null>(
    null
  );
  const [error, setError] = useState<string | null>(null);

  // Load saved snippets from backend API
  useEffect(() => {
    loadUserSnippets();
  }, [languageId]);

  const loadUserSnippets = async () => {
    try {
      const snippets = await contentService.getSnippetsForLanguage(languageId);
      const customSnippets = snippets
        .filter(snippet => snippet.tags && snippet.tags.includes('custom'))
        .map(snippet => ({
          id: snippet.id,
          code: snippet.sourceCode,
          language: languageId,
          name: snippet.title,
          createdAt: new Date(), // Use current date since SnippetData doesn't have createdAt
          userId: authService.getCurrentUser()?.id,
          tags: snippet.tags,
          difficulty: snippet.difficulty,
        }));
      setSavedSnippets(customSnippets);
    } catch (err) {
      console.error('Failed to load user snippets:', err);
      setError('Failed to load saved snippets');
    }
  };

  const handleSaveSnippet = useCallback(async () => {
    if (!customCode.trim()) {
      setError('Please enter some code');
      return;
    }

    if (!snippetName.trim()) {
      setError('Please enter a name for your snippet');
      return;
    }

    try {
      setError(null);

      // Save snippet via API
      const newSnippet = {
        id: Date.now().toString(),
        title: snippetName,
        sourceCode: customCode,
        languageId: selectedLanguage,
        tags: ['custom'],
        difficulty: 3,
      };

      // Use contentService to create snippet (this will integrate with backend)
      await contentService.initialize(); // Ensure content is loaded

      // For now, we'll add to local state until the backend integration is complete
      // In a full implementation, this would call an API endpoint
      const customSnippet: CustomSnippet = {
        id: newSnippet.id,
        code: newSnippet.sourceCode,
        language: selectedLanguage,
        name: newSnippet.title,
        createdAt: new Date(),
        userId: authService.getCurrentUser()?.id,
        tags: newSnippet.tags,
        difficulty: newSnippet.difficulty,
      };

      setSavedSnippets(prev => [...prev, customSnippet]);

      // Reset form
      setCustomCode('');
      setSnippetName('');
      setShowEditor(false);
    } catch (err) {
      console.error('Failed to save snippet:', err);
      setError('Failed to save snippet');
    }
  }, [customCode, snippetName, selectedLanguage]);

  const handleStartPractice = useCallback((snippet: CustomSnippet) => {
    setActiveSnippet(snippet);
  }, []);

  const handleDeleteSnippet = useCallback(
    (id: string) => {
      const updated = savedSnippets.filter(s => s.id !== id);
      setSavedSnippets(updated);
      try {
        localStorage.setItem('customSnippets', JSON.stringify(updated));
      } catch (err) {
        console.error('Failed to delete snippet:', err);
      }
    },
    [savedSnippets]
  );

  const handleComplete = useCallback(
    (result?: SessionResult) => {
      setActiveSnippet(null);
      if (onComplete && result) {
        onComplete(result);
      }
    },
    [onComplete]
  );

  if (activeSnippet) {
    return (
      <div className="custom-snippets-practice">
        <div className="practice-header">
          <h3>{activeSnippet.name}</h3>
          <button onClick={() => setActiveSnippet(null)} className="btn-back">
            ← Back to Snippets
          </button>
        </div>
        <MonacoTypingInterface
          targetText={activeSnippet.code}
          languageId={activeSnippet.language}
          currentText=""
          onTextChange={() => {
            // Monaco will handle text changes internally
          }}
          onComplete={() => {
            // Monaco calls onComplete when typing is complete
            // We need to create a session result and call handleComplete
            handleComplete();
          }}
          theme="dark"
          fontSize={16}
          lineHeight={1.6}
          keyboardLayout="qwerty"
          className="custom-snippets-editor"
        />
      </div>
    );
  }

  if (showEditor) {
    return (
      <div className="custom-snippets-editor">
        <div className="editor-header">
          <h2>Create Custom Snippet</h2>
          <button onClick={() => setShowEditor(false)} className="btn-close">
            ✕
          </button>
        </div>

        <div className="editor-form">
          <div className="form-group">
            <label htmlFor="snippet-name">Snippet Name</label>
            <input
              id="snippet-name"
              type="text"
              value={snippetName}
              onChange={e => setSnippetName(e.target.value)}
              placeholder="e.g., React Component, Python Function..."
              className="form-input"
            />
          </div>

          <div className="form-group">
            <label htmlFor="snippet-language">Language</label>
            <select
              id="snippet-language"
              value={selectedLanguage}
              onChange={e => setSelectedLanguage(e.target.value as Language)}
              className="form-select"
            >
              <option value="javascript">JavaScript</option>
              <option value="python">Python</option>
              <option value="cpp">C++</option>
              <option value="rust">Rust</option>
              <option value="yaml">YAML</option>
            </select>
          </div>

          <div className="form-group">
            <label htmlFor="snippet-code">Code</label>
            <textarea
              id="snippet-code"
              value={customCode}
              onChange={e => setCustomCode(e.target.value)}
              placeholder="Paste or type your code here..."
              className="form-textarea"
              rows={15}
              spellCheck={false}
            />
          </div>

          {error && <div className="error-message">{error}</div>}

          <div className="form-actions">
            <button onClick={handleSaveSnippet} className="btn-save">
              Save Snippet
            </button>
            <button onClick={() => setShowEditor(false)} className="btn-cancel">
              Cancel
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="custom-snippets-mode">
      <div className="snippets-header">
        <div>
          <h2>Custom Snippets</h2>
          <p>Practice with your own code</p>
        </div>
        <div className="header-actions">
          <button onClick={() => setShowEditor(true)} className="btn-create">
            + Create New Snippet
          </button>
          <button onClick={onExit} className="btn-exit">
            Exit
          </button>
        </div>
      </div>

      {savedSnippets.length === 0 ? (
        <div className="empty-state">
          <div className="empty-icon">📝</div>
          <h3>No Custom Snippets Yet</h3>
          <p>
            Create your first custom snippet to practice with your own code. You
            can paste code from your projects or create practice exercises.
          </p>
          <button
            onClick={() => setShowEditor(true)}
            className="btn-create-large"
          >
            Create Your First Snippet
          </button>
        </div>
      ) : (
        <div className="snippets-grid">
          {savedSnippets.map(snippet => (
            <div key={snippet.id} className="snippet-card">
              <div className="snippet-header">
                <h3>{snippet.name}</h3>
                <span className="snippet-language">
                  {snippet.language.toUpperCase()}
                </span>
              </div>
              <pre className="snippet-preview">
                <code>{snippet.code.substring(0, 150)}...</code>
              </pre>
              <div className="snippet-meta">
                <span className="snippet-date">
                  {snippet.createdAt.toLocaleDateString()}
                </span>
                <span className="snippet-length">
                  {snippet.code.length} chars
                </span>
              </div>
              <div className="snippet-actions">
                <button
                  onClick={() => handleStartPractice(snippet)}
                  className="btn-practice"
                >
                  Practice
                </button>
                <button
                  onClick={() => handleDeleteSnippet(snippet.id)}
                  className="btn-delete"
                >
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default CustomSnippetsMode;
