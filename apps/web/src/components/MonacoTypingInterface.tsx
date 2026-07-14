import React, { useRef, useEffect, useState, useCallback } from 'react';
import Editor, { Monaco } from '@monaco-editor/react';
import KeyboardLayoutService, {
  KeyboardLayout,
} from '../services/KeyboardLayoutService';
import { useAccessibility } from './AccessibilitySettings';
interface MonacoTypingInterfaceProps {
  targetText: string;
  languageId: string;
  currentText: string;
  onTextChange: (text: string) => void;
  onComplete: () => void;
  theme?: 'light' | 'dark' | 'solarized';
  fontSize?: number;
  lineHeight?: number;
  keyboardLayout?: 'qwerty' | 'azerty' | 'qwertz' | 'colemak' | 'dvorak';
  readOnly?: boolean;
  showLineNumbers?: boolean;
  className?: string;
  onFontSizeChange?: (size: number) => void;
  onLineHeightChange?: (height: number) => void;
  enableFontControls?: boolean;
  enableLineHeightControls?: boolean;
}

const MonacoTypingInterface: React.FC<MonacoTypingInterfaceProps> = ({
  targetText,
  languageId,
  currentText,
  onTextChange: _onTextChange,
  onComplete: _onComplete,
  theme = 'dark',
  fontSize = 16,
  lineHeight = 1.6,
  keyboardLayout = 'qwerty',
  readOnly = true,
  showLineNumbers = false,
  className = '',
}) => {
  const editorRef = useRef<any>(null);
  const overlayRef = useRef<HTMLDivElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [monaco, setMonaco] = useState<Monaco | null>(null);
  const { preferences } = useAccessibility();

  // Use accessibility preferences for font and line height
  const fontSizeSetting = preferences.fontSize ?? fontSize;
  const lineHeightSetting = preferences.lineHeight ?? lineHeight;
  const [currentKeyboardLayout, setCurrentKeyboardLayout] =
    useState<KeyboardLayout | null>(null);

  // Load keyboard layout
  useEffect(() => {
    const layout = KeyboardLayoutService.getLayout(keyboardLayout);
    setCurrentKeyboardLayout(layout);
  }, [keyboardLayout]);

  // Announcement and focus state
  const [announcements, setAnnouncements] = useState<string[]>([]);
  const [isTypingMode, setIsTypingMode] = useState(false);

  const announce = useCallback(
    (message: string) => {
      if (preferences.screenReaderOptimizations) {
        setAnnouncements(prev => [...prev, message]);
      }
    },
    [preferences.screenReaderOptimizations]
  );

  // Focus manager and keyboard navigation
  const focusEditor = useCallback(() => {
    if (editorRef.current) {
      editorRef.current.focus();
      setIsTypingMode(true);
      announce('Entered typing mode. Start typing to begin practice.');
    }
  }, [announce]);

  const focusContainer = useCallback(() => {
    containerRef.current?.focus();
    setIsTypingMode(false);
    announce('Exited typing mode');
  }, [announce]);

  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      if (isTypingMode && event.key.length === 1) {
        return;
      }

      switch (event.key) {
        case 'Enter':
        case ' ':
          event.preventDefault();
          focusEditor();
          break;
        case 'Escape':
          if (isTypingMode) {
            event.preventDefault();
            focusContainer();
          }
          break;
        case 'ArrowUp':
        case 'ArrowDown':
        case 'ArrowLeft':
        case 'ArrowRight':
          if (isTypingMode) {
            return;
          }
          break;
      }
    },
    [isTypingMode, focusEditor, focusContainer]
  );

  // Announce typing progress to screen readers
  useEffect(() => {
    if (currentText.length > 0 && currentText.length % 5 === 0) {
      const progress = Math.round(
        (currentText.length / targetText.length) * 100
      );
      announce(
        `Progress: ${progress}% complete. ${currentText.length} of ${targetText.length} characters typed.`
      );
    }
  }, [currentText.length, targetText.length, announce]);

  // Announce typing errors to screen readers
  useEffect(() => {
    if (currentText.length > 0) {
      const lastChar = currentText[currentText.length - 1];
      const expectedChar = targetText[currentText.length - 1];
      if (lastChar !== expectedChar) {
        announce(
          `Error: expected ${expectedChar}, got ${lastChar} at position ${currentText.length}`
        );
      }
    }
  }, [currentText, targetText, announce]);

  // Announce keyboard layout changes to screen readers
  useEffect(() => {
    if (currentKeyboardLayout) {
      announce(`Keyboard layout: ${currentKeyboardLayout.displayName}`);
    }
  }, [currentKeyboardLayout, announce]);

  // Update editor options when accessibility font/line-height preferences change
  useEffect(() => {
    if (editorRef.current) {
      editorRef.current.updateOptions({
        fontSize: fontSizeSetting,
        lineHeight: lineHeightSetting,
      });
    }
  }, [fontSizeSetting, lineHeightSetting]);

  // Capture Escape while the editor is focused so focus returns to the container
  useEffect(() => {
    if (!isTypingMode || !editorRef.current) return;

    const node = editorRef.current.getContainerDomNode?.();
    if (!node) return;

    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        focusContainer();
      }
    };

    node.addEventListener('keydown', handleKeyDown);
    return () => node.removeEventListener('keydown', handleKeyDown);
  }, [isTypingMode, focusContainer]);

  // Apply theme to Monaco Editor
  const applyTheme = useCallback(
    (monacoInstance: Monaco, themeName: string) => {
      if (!monacoInstance) return;

      const isDark = themeName === 'dark' || themeName === 'solarized';

      // Solarized color palette
      const solarizedColors = {
        base03: '#002b36', // background
        base02: '#073642', // background highlights
        base01: '#586e75', // comments / secondary content
        base00: '#657b83', // body text / default code
        base0: '#839496', // body text / default code (emphasized)
        base1: '#93a1a1', // comments / secondary content (emphasized)
        base2: '#eee8d5', // background highlights (emphasized)
        base3: '#fdf6e3', // background
        yellow: '#b58900', // yellow
        orange: '#cb4b16', // orange
        red: '#dc322f', // red
        magenta: '#d33682', // magenta
        violet: '#6c71c4', // violet
        blue: '#268bd2', // blue
        cyan: '#2aa198', // cyan
        green: '#859900', // green
      };

      let themeRules: any[] = [];
      let themeColors: any = {};

      switch (themeName) {
        case 'solarized':
          themeRules = [
            { token: 'comment', foreground: solarizedColors.base01 },
            { token: 'keyword', foreground: solarizedColors.blue },
            { token: 'string', foreground: solarizedColors.cyan },
            { token: 'number', foreground: solarizedColors.cyan },
            { token: 'type', foreground: solarizedColors.yellow },
            { token: 'class', foreground: solarizedColors.yellow },
            { token: 'function', foreground: solarizedColors.blue },
            { token: 'variable', foreground: solarizedColors.blue },
            { token: 'constant', foreground: solarizedColors.cyan },
          ];
          themeColors = {
            'editor.background': solarizedColors.base03,
            'editor.foreground': solarizedColors.base0,
            'editor.lineHighlightBackground': solarizedColors.base02,
            'editor.selectionBackground': solarizedColors.base2,
            'editorCursor.foreground': solarizedColors.base1,
            'editorWhitespace.foreground': solarizedColors.base01,
            'editorLineNumber.foreground': solarizedColors.base01,
            'editorLineNumber.activeForeground': solarizedColors.base1,
          };
          break;

        case 'dark':
          themeRules = [
            { token: 'comment', foreground: '6A9955' },
            { token: 'keyword', foreground: '569CD6' },
            { token: 'string', foreground: 'CE9178' },
            { token: 'number', foreground: 'B5CEA8' },
          ];
          themeColors = {
            'editor.background': '#1e1e1e',
            'editor.foreground': '#d4d4d4',
            'editor.lineHighlightBackground': '#2d2d2d',
            'editor.selectionBackground': '#264f78',
            'editorCursor.foreground': '#ffffff',
            'editorWhitespace.foreground': '#404040',
          };
          break;

        case 'light':
          themeRules = [
            { token: 'comment', foreground: '008000' },
            { token: 'keyword', foreground: '0000FF' },
            { token: 'string', foreground: 'A31515' },
            { token: 'number', foreground: '098658' },
          ];
          themeColors = {
            'editor.background': '#ffffff',
            'editor.foreground': '#000000',
            'editor.lineHighlightBackground': '#f0f0f0',
            'editor.selectionBackground': '#add6ff',
            'editorCursor.foreground': '#000000',
            'editorWhitespace.foreground': '#d3d3d3',
          };
          break;
      }

      // Define custom theme
      monacoInstance.editor.defineTheme('custom-theme', {
        base: isDark ? 'vs-dark' : 'vs',
        inherit: true,
        rules: themeRules,
        colors: themeColors,
      });

      monacoInstance.editor.setTheme('custom-theme');
    },
    []
  );

  // Handle Monaco Editor mount
  const handleEditorDidMount = useCallback(
    (editor: any, monacoInstance: Monaco) => {
      editorRef.current = editor;
      setMonaco(monacoInstance);

      // Configure editor options
      editor.updateOptions({
        fontSize: fontSizeSetting,
        lineHeight: lineHeightSetting,
        fontFamily: 'JetBrains Mono, Consolas, monospace',
        minimap: { enabled: false },
        scrollBeyondLastLine: false,
        wordWrap: 'on',
        lineNumbers: showLineNumbers ? 'on' : 'off',
        glyphMargin: false,
        folding: false,
        lineDecorationsWidth: 0,
        lineNumbersMinChars: 0,
        renderLineHighlight: 'none',
        scrollbar: {
          vertical: 'hidden',
          horizontal: 'hidden',
          handleMouseWheel: false,
        },
        overviewRulerLanes: 0,
        hideCursorInOverviewRuler: true,
        overviewRulerBorder: false,
        bracketPairColorization: { enabled: true },
        guides: {
          bracketPairs: false,
          indentation: false,
        },
        smoothScrolling: true,
        cursorBlinking: 'smooth',
        renderWhitespace: 'selection',
        renderControlCharacters: false,
        fontLigatures: true,
      });

      // Make editor read-only
      if (readOnly) {
        editor.updateOptions({
          readOnly: true,
        });
      }

      // Apply theme
      applyTheme(monacoInstance, theme);

      // Disable keyboard shortcuts that might interfere
      editor.addCommand(
        monacoInstance.KeyMod.CtrlCmd | monacoInstance.KeyCode.KeyS,
        () => {}
      );
      editor.addCommand(
        monacoInstance.KeyMod.CtrlCmd | monacoInstance.KeyCode.KeyP,
        () => {}
      );
    },
    [
      showLineNumbers,
      readOnly,
      theme,
      applyTheme,
      fontSizeSetting,
      lineHeightSetting,
    ]
  );

  // Update theme when it changes
  useEffect(() => {
    if (monaco) {
      applyTheme(monaco, theme);
    }
  }, [monaco, theme, applyTheme]);

  // Handle text changes from external source (typing)
  useEffect(() => {
    if (editorRef.current && monaco) {
      const editor = editorRef.current;
      const model = editor.getModel();

      if (model && model.getValue() !== currentText) {
        // Update the editor content to reflect typing progress
        const lines = targetText.split('\n');
        let currentPos = 0;

        for (
          let i = 0;
          i < lines.length && currentPos < currentText.length;
          i++
        ) {
          const lineLength = lines[i].length;
          if (currentPos + lineLength >= currentText.length) {
            // We're in the middle of this line
            const position = model.getPositionAt(
              currentPos + currentText.length - currentPos
            );
            editor.setPosition(position);
            break;
          }
          currentPos += lineLength + 1; // +1 for newline
        }
      }
    }
  }, [currentText, targetText, monaco]);

  // Render typing overlay with character-by-character feedback
  const renderTypingOverlay = useCallback(() => {
    if (!targetText) return null;

    return (
      <div
        ref={overlayRef}
        data-testid="typing-target-text"
        className={`typing-overlay ${className}`}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          pointerEvents: 'none',
          zIndex: 10,
          fontSize: `${fontSizeSetting}px`,
          lineHeight: lineHeightSetting,
          fontFamily: 'JetBrains Mono, Consolas, monospace',
          whiteSpace: 'pre-wrap',
          wordWrap: 'break-word',
          overflowWrap: 'break-word',
        }}
      >
        {targetText.split('').map((char, index) => {
          const isTyped = index < currentText.length;
          const isCorrect = isTyped && currentText[index] === char;
          const isCurrent = index === currentText.length;

          let className = 'typing-char';

          if (isCurrent) {
            className += ' typing-char-current';
          } else if (isTyped) {
            if (isCorrect) {
              className += ' typing-char-correct';
            } else {
              className += ' typing-char-incorrect';
            }
          } else {
            className += ' typing-char-pending';
          }

          return (
            <span key={index} className={className}>
              {char === '\n' ? '↵\n' : char}
            </span>
          );
        })}
        {/* Progress indicator */}
        <div
          className="typing-progress"
          style={{
            position: 'absolute',
            bottom: '10px',
            right: '10px',
            background: 'rgba(0, 0, 0, 0.7)',
            color: 'white',
            padding: '4px 8px',
            borderRadius: '4px',
            fontSize: '12px',
          }}
        >
          {Math.round((currentText.length / targetText.length) * 100)}%
        </div>
      </div>
    );
  }, [targetText, currentText, fontSizeSetting, lineHeightSetting, className]);

  return (
    <div
      ref={containerRef}
      className="monaco-typing-container"
      style={{ position: 'relative', width: '100%', height: '100%' }}
      onKeyDown={handleKeyDown}
      tabIndex={0}
    >
      <Editor
        height="100%"
        language={languageId}
        value={targetText}
        onMount={handleEditorDidMount}
        options={{
          readOnly: true,
          minimap: { enabled: false },
          scrollBeyondLastLine: false,
          wordWrap: 'on',
          lineNumbers: showLineNumbers ? 'on' : 'off',
          glyphMargin: false,
          folding: false,
          lineDecorationsWidth: 0,
          lineNumbersMinChars: 0,
          renderLineHighlight: 'none',
          scrollbar: {
            vertical: 'hidden',
            horizontal: 'hidden',
          },
        }}
        theme={theme === 'dark' ? 'vs-dark' : 'light'}
        aria-label={`Code typing practice in ${languageId}`}
        aria-describedby="typing-instructions"
      />

      {/* ARIA live region for screen reader announcements */}
      <div
        role="status"
        aria-live="polite"
        aria-atomic="true"
        className="sr-only"
        style={{
          position: 'absolute',
          left: '-10000px',
          width: '1px',
          height: '1px',
          overflow: 'hidden',
        }}
      >
        {announcements.length > 0 && (
          <div key={announcements.length}>
            {announcements[announcements.length - 1]}
          </div>
        )}
      </div>

      {/* Hidden instructions for screen readers */}
      <div id="typing-instructions" className="sr-only">
        Type the code shown above. Use Tab to reach the editor, then Enter to
        focus it and Escape to return. Your progress and any errors will be
        announced automatically.
      </div>

      {/* Custom typing overlay */}
      {renderTypingOverlay()}
    </div>
  );
};

export default MonacoTypingInterface;
