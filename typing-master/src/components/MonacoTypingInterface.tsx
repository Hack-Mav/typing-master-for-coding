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
  const [monaco, setMonaco] = useState<Monaco | null>(null);
  const { preferences } = useAccessibility();

  // Use accessibility preferences for font and line height
  const [fontSizeSetting] = useState(preferences.fontSize || fontSize);
  const [lineHeightSetting] = useState(preferences.lineHeight || lineHeight);
  const [currentKeyboardLayout, setCurrentKeyboardLayout] =
    useState<KeyboardLayout | null>(null);

  // Load keyboard layout
  useEffect(() => {
    const layout = KeyboardLayoutService.getLayout(keyboardLayout);
    setCurrentKeyboardLayout(layout);
  }, [keyboardLayout]);

  // Keyboard navigation state
  const [focusedElement, setFocusedElement] = useState<string | null>(null);
  const [announcements, setAnnouncements] = useState<string[]>([]);
  const [isTypingMode, setIsTypingMode] = useState(false);

  // Enhanced keyboard navigation
  const handleKeyDown = useCallback(
    (event: React.KeyboardEvent) => {
      // Don't interfere with typing when in typing mode
      if (isTypingMode && event.key.length === 1) {
        return;
      }

      switch (event.key) {
        case 'Tab':
          event.preventDefault();
          // Handle tab navigation between focusable elements
          const focusableElements = [
            'editor',
            'theme-button',
            'summary-button',
            'exit-button',
          ];
          const currentIndex = focusedElement
            ? focusableElements.indexOf(focusedElement)
            : 0;
          const nextIndex = event.shiftKey
            ? (currentIndex - 1 + focusableElements.length) %
              focusableElements.length
            : (currentIndex + 1) % focusableElements.length;
          setFocusedElement(focusableElements[nextIndex]);
          setAnnouncements(prev => [
            ...prev,
            `Navigated to ${focusableElements[nextIndex]}`,
          ]);
          break;
        case 'Enter':
        case ' ':
          if (focusedElement === 'editor') {
            event.preventDefault();
            setIsTypingMode(true);
            setAnnouncements(prev => [
              ...prev,
              'Entered typing mode. Start typing to begin practice.',
            ]);
          } else if (focusedElement === 'theme-button') {
            event.preventDefault();
            setAnnouncements(prev => [...prev, 'Theme button activated']);
          } else if (focusedElement === 'summary-button') {
            event.preventDefault();
            setAnnouncements(prev => [...prev, 'Summary button activated']);
          } else if (focusedElement === 'exit-button') {
            event.preventDefault();
            setAnnouncements(prev => [...prev, 'Exit button activated']);
          }
          break;
        case 'Escape':
          if (isTypingMode) {
            setIsTypingMode(false);
            setAnnouncements(prev => [...prev, 'Exited typing mode']);
          } else {
            setAnnouncements(prev => [...prev, 'Escape pressed']);
          }
          event.preventDefault();
          break;
        case 'ArrowUp':
        case 'ArrowDown':
        case 'ArrowLeft':
        case 'ArrowRight':
          // Allow normal cursor movement in editor
          if (focusedElement === 'editor') {
            return;
          }
          break;
      }
    },
    [focusedElement, isTypingMode]
  );

  // Enhanced announcements for screen readers
  useEffect(() => {
    if (currentText.length > 0 && currentText.length % 5 === 0) {
      const progress = Math.round(
        (currentText.length / targetText.length) * 100
      );
      setAnnouncements(prev => [
        ...prev,
        `Progress: ${progress}% complete. ${currentText.length} of ${targetText.length} characters typed.`,
      ]);
    }
  }, [currentText.length, targetText.length]);

  // Enhanced error announcements
  useEffect(() => {
    if (currentText.length > 0) {
      const lastChar = currentText[currentText.length - 1];
      const expectedChar = targetText[currentText.length - 1];
      if (lastChar !== expectedChar) {
        setAnnouncements(prev => [
          ...prev,
          `Error: expected ${expectedChar}, got ${lastChar} at position ${currentText.length}`,
        ]);
      }
    }
  }, [currentText, targetText]);

  // Keyboard layout indicator for screen readers
  useEffect(() => {
    if (currentKeyboardLayout) {
      setAnnouncements(prev => [
        ...prev,
        `Keyboard layout: ${currentKeyboardLayout.displayName}`,
      ]);
    }
  }, [currentKeyboardLayout]);

  // High contrast mode support
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-contrast: high)');
    const handleChange = () => {
      if (mediaQuery.matches) {
        // Apply high contrast theme adjustments
        document.body.classList.add('high-contrast');
      } else {
        document.body.classList.remove('high-contrast');
      }
    };

    handleChange();
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  // Reduced motion support
  useEffect(() => {
    const mediaQuery = window.matchMedia('(prefers-reduced-motion: reduce)');
    const handleChange = () => {
      if (mediaQuery.matches) {
        document.body.classList.add('reduced-motion');
      } else {
        document.body.classList.remove('reduced-motion');
      }
    };

    handleChange();
    mediaQuery.addEventListener('change', handleChange);
    return () => mediaQuery.removeEventListener('change', handleChange);
  }, []);

  // Handle text changes and announce to screen readers
  useEffect(() => {
    if (currentText.length > 0 && currentText.length % 10 === 0) {
      const progress = Math.round(
        (currentText.length / targetText.length) * 100
      );
      setAnnouncements(prev => [...prev, `Progress: ${progress}% complete`]);
    }
  }, [currentText.length, targetText.length]);

  // Handle errors and announce to screen readers
  useEffect(() => {
    if (currentText.length > 0) {
      const lastChar = currentText[currentText.length - 1];
      const expectedChar = targetText[currentText.length - 1];
      if (lastChar !== expectedChar) {
        setAnnouncements(prev => [
          ...prev,
          `Error: expected ${expectedChar}, got ${lastChar}`,
        ]);
      }
    }
  }, [currentText, targetText]);

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
          const hasError = isTyped && !isCorrect;

          let className = 'typing-char';
          let beforeContent = '';
          let afterContent = '';

          if (isCurrent) {
            className += ' typing-char-current';
            beforeContent = '█';
          } else if (isTyped) {
            if (isCorrect) {
              className += ' typing-char-correct';
            } else {
              className += ' typing-char-incorrect';
            }
          } else {
            className += ' typing-char-pending';
          }

          // Add error indicators
          if (hasError) {
            afterContent = '✗';
          }

          return (
            <span key={index} className={className}>
              {beforeContent}
              {char === '\n'
                ? '↵\n'
                : char === ' '
                  ? '·'
                  : char === '\t'
                    ? '→   '
                    : char}
              {afterContent}
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
      className="monaco-typing-container"
      style={{ position: 'relative', width: '100%', height: '100%' }}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
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
        {announcements.map((announcement, index) => (
          <div key={index}>{announcement}</div>
        ))}
      </div>

      {/* Hidden instructions for screen readers */}
      <div id="typing-instructions" className="sr-only">
        Type the code shown above. Use Tab to navigate and Enter to start
        typing. Your progress and any errors will be announced automatically.
      </div>

      {/* Custom typing overlay */}
      {renderTypingOverlay()}
    </div>
  );
};

export default MonacoTypingInterface;
