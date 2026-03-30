# Typing Master for Coding - VS Code Extension

Practice typing real code with syntax-aware guidance directly in Visual Studio Code.

## Features

- **Embedded Practice Sessions**: Practice typing code without leaving your editor
- **Multiple Practice Modes**: Timed drills, accuracy mode, and zen mode
- **Real-time Metrics**: See your CPM, accuracy, and progress in real-time
- **Language Support**: Practice with JavaScript, TypeScript, Python, Java, C++, Rust, and YAML
- **Current File Practice**: Practice typing code from your currently open file
- **Selection Practice**: Practice typing selected code snippets
- **Leaderboard Integration**: Compete with other developers
- **Seamless Integration**: Works directly within VS Code with native UI

## Commands

- `Start Typing Practice`: Begin a new practice session with default code
- `Practice Current File`: Practice typing the code in your currently active editor
- `Practice Selection`: Practice typing the currently selected text
- `Show Leaderboard`: View the global leaderboard

## Configuration

The extension can be configured through VS Code settings:

- `typing-master.apiUrl`: API URL for the Typing Master backend
- `typing-master.practiceMode`: Default practice mode (timed, accuracy, zen)
- `typing-master.duration`: Practice duration in minutes (for timed mode)
- `typing-master.showRealTimeMetrics`: Show real-time typing metrics during practice
- `typing-master.autoSaveResults`: Automatically save practice results to leaderboards

## Installation

1. Install the extension from the VS Code Marketplace
2. Open the Command Palette (Ctrl+Shift+P or Cmd+Shift+P)
3. Type "Typing Master" to see available commands
4. Choose a practice mode and start typing!

## Usage

### Starting a Practice Session

1. Use the command palette and select "Start Typing Practice"
2. Choose your preferred programming language
3. The practice interface will open in a new panel
4. Start typing the code shown

### Practicing Current File

1. Open a code file in VS Code
2. Use the command palette and select "Practice Current File"
3. The practice interface will open with your file's code

### Practicing Selection

1. Select code in your editor
2. Right-click and choose "Practice Selection"
3. The practice interface will open with the selected code

## Supported Languages

- JavaScript
- TypeScript
- Python
- Java
- C++
- Rust
- YAML

## Requirements

- VS Code 1.74.0 or higher
- Internet connection for backend services

## Extension Settings

This extension contributes the following settings:

* `typing-master.apiUrl`: Default: `https://api.typing-master-for-coding.com`
* `typing-master.practiceMode`: Default: `timed`
* `typing-master.duration`: Default: `3`
* `typing-master.showRealTimeMetrics`: Default: `true`
* `typing-master.autoSaveResults`: Default: `true`

## Development

For development and testing:

1. Clone this repository
2. Run `npm install` in the extension directory
3. Run `npm run compile` to build the extension
4. Press F5 to open a new Extension Development Host window
5. Test the extension in the development host

## Contributing

Contributions are welcome! Please read the contributing guidelines and submit pull requests.

## License

This extension is licensed under the MIT License.

## Support

For issues and feature requests, please use the GitHub issues page.

---

**Happy typing!** 🎹⌨️
