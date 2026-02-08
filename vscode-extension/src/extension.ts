import * as vscode from 'vscode';
import axios from 'axios';

export function activate(context: vscode.ExtensionContext) {
    console.log('Typing Master for Coding extension is now active!');

    // Get configuration
    const config = vscode.workspace.getConfiguration('typing-master');
    const apiUrl = config.get<string>('apiUrl', 'https://api.typing-master-for-coding.com');
    const practiceMode = config.get<string>('practiceMode', 'timed');
    const duration = config.get<number>('duration', 3);
    const showRealTimeMetrics = config.get<boolean>('showRealTimeMetrics', true);
    const autoSaveResults = config.get<boolean>('autoSaveResults', true);

    // Register commands
    const startPracticeCommand = vscode.commands.registerCommand('typing-master.startPractice', () => {
        startPracticeSession(context, apiUrl, practiceMode, duration, showRealTimeMetrics, autoSaveResults);
    });

    const practiceCurrentFileCommand = vscode.commands.registerCommand('typing-master.practiceCurrentFile', () => {
        practiceCurrentFile(context, apiUrl, practiceMode, duration, showRealTimeMetrics, autoSaveResults);
    });

    const practiceSelectionCommand = vscode.commands.registerCommand('typing-master.practiceSelection', () => {
        practiceSelection(context, apiUrl, practiceMode, duration, showRealTimeMetrics, autoSaveResults);
    });

    const showLeaderboardCommand = vscode.commands.registerCommand('typing-master.showLeaderboard', () => {
        showLeaderboard(context, apiUrl);
    });

    const shareToGitHubCommand = vscode.commands.registerCommand('typing-master.shareToGitHub', () => {
        shareToGitHub(context, apiUrl);
    });

    const syncWithGitHubCommand = vscode.commands.registerCommand('typing-master.syncWithGitHub', () => {
        syncWithGitHub(context, apiUrl);
    });

    // Add commands to context
    context.subscriptions.push(
        startPracticeCommand,
        practiceCurrentFileCommand,
        practiceSelectionCommand,
        showLeaderboardCommand,
        shareToGitHubCommand,
        syncWithGitHubCommand
    );

    // Set context for activity bar
    vscode.commands.executeCommand('setContext', 'typing-master.active', true);
}

export function deactivate() {
    console.log('Typing Master for Coding extension is now deactivated');
}

async function startPracticeSession(
    context: vscode.ExtensionContext,
    apiUrl: string,
    practiceMode: string,
    duration: number,
    showRealTimeMetrics: boolean,
    autoSaveResults: boolean
) {
    try {
        // Show language selection
        const languages = ['javascript', 'python', 'typescript', 'java', 'cpp', 'rust', 'yaml'];
        const selectedLanguage = await vscode.window.showQuickPick(languages, {
            placeHolder: 'Select a programming language to practice'
        });

        if (!selectedLanguage) {
            return;
        }

        // Create embedded session
        const response = await axios.post(`${apiUrl}/api/v1/integrations/embedded/session`, {
            source_code: getDefaultCodeForLanguage(selectedLanguage),
            language: selectedLanguage,
            mode: practiceMode,
            user_id: context.globalState.get('userId', 'anonymous')
        });

        const { session_id, embed_url } = response.data;

        // Show practice panel
        const panel = vscode.window.createWebviewPanel(
            'typingPractice',
            'Typing Practice',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = getPracticeWebviewContent(embed_url, session_id, showRealTimeMetrics);

        // Handle messages from webview
        panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'practiceCompleted':
                        if (autoSaveResults) {
                            await savePracticeResults(message.results, apiUrl);
                        }
                        vscode.window.showInformationMessage(
                            `Practice completed! Score: ${message.results.composite_score}`
                        );
                        break;
                    case 'practiceError':
                        vscode.window.showErrorMessage(`Practice error: ${message.error}`);
                        break;
                }
            },
            undefined,
            context.subscriptions
        );

    } catch (error) {
        vscode.window.showErrorMessage(`Failed to start practice session: ${error}`);
    }
}

async function practiceCurrentFile(
    context: vscode.ExtensionContext,
    apiUrl: string,
    practiceMode: string,
    duration: number,
    showRealTimeMetrics: boolean,
    autoSaveResults: boolean
) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor found');
        return;
    }

    const document = editor.document;
    const sourceCode = document.getText();
    const language = detectLanguageFromDocument(document);

    try {
        const response = await axios.post(`${apiUrl}/api/v1/integrations/embedded/session`, {
            source_code: sourceCode,
            language: language,
            mode: practiceMode,
            user_id: context.globalState.get('userId', 'anonymous')
        });

        const { session_id, embed_url } = response.data;

        const panel = vscode.window.createWebviewPanel(
            'typingPractice',
            'Typing Practice - Current File',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = getPracticeWebviewContent(embed_url, session_id, showRealTimeMetrics);

        panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'practiceCompleted':
                        if (autoSaveResults) {
                            await savePracticeResults(message.results, apiUrl);
                        }
                        vscode.window.showInformationMessage(
                            `Practice completed! Score: ${message.results.composite_score}`
                        );
                        break;
                    case 'practiceError':
                        vscode.window.showErrorMessage(`Practice error: ${message.error}`);
                        break;
                }
            },
            undefined,
            context.subscriptions
        );

    } catch (error) {
        vscode.window.showErrorMessage(`Failed to start practice session: ${error}`);
    }
}

async function practiceSelection(
    context: vscode.ExtensionContext,
    apiUrl: string,
    practiceMode: string,
    duration: number,
    showRealTimeMetrics: boolean,
    autoSaveResults: boolean
) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor found');
        return;
    }

    const selection = editor.selection;
    if (selection.isEmpty) {
        vscode.window.showWarningMessage('No text selected');
        return;
    }

    const sourceCode = editor.document.getText(selection);
    const language = detectLanguageFromDocument(editor.document);

    try {
        const response = await axios.post(`${apiUrl}/api/v1/integrations/embedded/session`, {
            source_code: sourceCode,
            language: language,
            mode: practiceMode,
            user_id: context.globalState.get('userId', 'anonymous')
        });

        const { session_id, embed_url } = response.data;

        const panel = vscode.window.createWebviewPanel(
            'typingPractice',
            'Typing Practice - Selection',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = getPracticeWebviewContent(embed_url, session_id, showRealTimeMetrics);

        panel.webview.onDidReceiveMessage(
            async message => {
                switch (message.command) {
                    case 'practiceCompleted':
                        if (autoSaveResults) {
                            await savePracticeResults(message.results, apiUrl);
                        }
                        vscode.window.showInformationMessage(
                            `Practice completed! Score: ${message.results.composite_score}`
                        );
                        break;
                    case 'practiceError':
                        vscode.window.showErrorMessage(`Practice error: ${message.error}`);
                        break;
                }
            },
            undefined,
            context.subscriptions
        );

    } catch (error) {
        vscode.window.showErrorMessage(`Failed to start practice session: ${error}`);
    }
}

async function showLeaderboard(context: vscode.ExtensionContext, apiUrl: string) {
    try {
        const response = await axios.get(`${apiUrl}/api/v1/leaderboard`);
        const leaderboard = response.data;

        const panel = vscode.window.createWebviewPanel(
            'leaderboard',
            'Typing Master Leaderboard',
            vscode.ViewColumn.One,
            {
                enableScripts: true,
                retainContextWhenHidden: true
            }
        );

        panel.webview.html = getLeaderboardWebviewContent(leaderboard);

    } catch (error) {
        vscode.window.showErrorMessage(`Failed to load leaderboard: ${error}`);
    }
}

async function savePracticeResults(results: any, apiUrl: string) {
    try {
        await axios.post(`${apiUrl}/api/v1/sessions/finalize`, results);
    } catch (error) {
        console.error('Failed to save practice results:', error);
    }
}

function detectLanguageFromDocument(document: vscode.TextDocument): string {
    const languageId = document.languageId;
    const languageMap: { [key: string]: string } = {
        'javascript': 'javascript',
        'typescript': 'typescript',
        'python': 'python',
        'java': 'java',
        'cpp': 'cpp',
        'c': 'cpp',
        'rust': 'rust',
        'yaml': 'yaml',
        'json': 'javascript'
    };
    return languageMap[languageId] || 'javascript';
}

function getDefaultCodeForLanguage(language: string): string {
    const defaults: { [key: string]: string } = {
        'javascript': 'function fibonacci(n) {\n    if (n <= 1) return n;\n    return fibonacci(n - 1) + fibonacci(n - 2);\n}',
        'python': 'def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n - 1) + fibonacci(n - 2)',
        'typescript': 'function fibonacci(n: number): number {\n    if (n <= 1) return n;\n    return fibonacci(n - 1) + fibonacci(n - 2);\n}',
        'java': 'public int fibonacci(int n) {\n    if (n <= 1) return n;\n    return fibonacci(n - 1) + fibonacci(n - 2);\n}',
        'cpp': 'int fibonacci(int n) {\n    if (n <= 1) return n;\n    return fibonacci(n - 1) + fibonacci(n - 2);\n}',
        'rust': 'fn fibonacci(n: i32) -> i32 {\n    if n <= 1 { return n; }\n    fibonacci(n - 1) + fibonacci(n - 2)\n}',
        'yaml': 'app:\n  name: typing-master\n  version: 1.0.0\n  config:\n    debug: true'
    };
    return defaults[language] || defaults['javascript'];
}

function getPracticeWebviewContent(embedUrl: string, sessionId: string, showRealTimeMetrics: boolean): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Typing Practice</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            background-color: var(--vscode-editor-background);
            color: var(--vscode-editor-foreground);
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .metrics {
            display: ${showRealTimeMetrics ? 'block' : 'none'};
            background-color: var(--vscode-editor-background);
            border: 1px solid var(--vscode-panel-border);
            border-radius: 4px;
            padding: 10px;
            margin-bottom: 20px;
        }
        .practice-area {
            background-color: var(--vscode-editor-background);
            border: 1px solid var(--vscode-panel-border);
            border-radius: 4px;
            padding: 20px;
            min-height: 400px;
        }
        .metric-item {
            display: inline-block;
            margin-right: 20px;
            font-weight: bold;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="metrics" id="metrics">
            <div class="metric-item">CPM: <span id="cpm">0</span></div>
            <div class="metric-item">Accuracy: <span id="accuracy">100%</span></div>
            <div class="metric-item">Time: <span id="time">0:00</span></div>
        </div>
        <div class="practice-area">
            <iframe src="${embedUrl}" width="100%" height="600" frameborder="0"></iframe>
        </div>
    </div>
    <script>
        const vscode = acquireVsCodeApi();
        
        // Listen for messages from the iframe
        window.addEventListener('message', event => {
            const message = event.data;
            
            switch (message.type) {
                case 'metrics':
                    updateMetrics(message.data);
                    break;
                case 'completed':
                    vscode.postMessage({
                        command: 'practiceCompleted',
                        results: message.data
                    });
                    break;
                case 'error':
                    vscode.postMessage({
                        command: 'practiceError',
                        error: message.data.error
                    });
                    break;
            }
        });
        
        function updateMetrics(metrics) {
            document.getElementById('cpm').textContent = metrics.cpm || 0;
            document.getElementById('accuracy').textContent = (metrics.accuracy || 100) + '%';
            document.getElementById('time').textContent = formatTime(metrics.time || 0);
        }
        
        function formatTime(seconds) {
            const minutes = Math.floor(seconds / 60);
            const secs = seconds % 60;
            return \`\${minutes}:\${secs.toString().padStart(2, '0')}\`;
        }
    </script>
</body>
</html>`;
}

function getLeaderboardWebviewContent(leaderboard: any): string {
    return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Leaderboard</title>
    <style>
        body {
            font-family: var(--vscode-font-family);
            background-color: var(--vscode-editor-background);
            color: var(--vscode-editor-foreground);
            margin: 0;
            padding: 20px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 10px;
            text-align: left;
            border-bottom: 1px solid var(--vscode-panel-border);
        }
        th {
            background-color: var(--vscode-editor-background);
            font-weight: bold;
        }
        .rank-1 { color: gold; }
        .rank-2 { color: silver; }
        .rank-3 { color: #cd7f32; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Typing Master Leaderboard</h1>
        <table>
            <thead>
                <tr>
                    <th>Rank</th>
                    <th>User</th>
                    <th>Score</th>
                    <th>CPM</th>
                    <th>Accuracy</th>
                    <th>Language</th>
                </tr>
            </thead>
            <tbody>
                ${leaderboard.map((entry: any, index: number) => `
                    <tr class="rank-${index + 1}">
                        <td>${index + 1}</td>
                        <td>${entry.user_handle || 'Anonymous'}</td>
                        <td>${entry.composite_score}</td>
                        <td>${entry.cpm}</td>
                        <td>${entry.accuracy}%</td>
                        <td>${entry.language}</td>
                    </tr>
                `).join('')}
            </tbody>
        </table>
    </div>
</body>
</html>`;
}

async function shareToGitHub(context: vscode.ExtensionContext, apiUrl: string) {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
        vscode.window.showWarningMessage('No active editor found');
        return;
    }

    const selection = editor.selection;
    const sourceCode = selection.isEmpty ? editor.document.getText() : editor.document.getText(selection);
    const language = detectLanguageFromDocument(editor.document);

    try {
        // Get GitHub token from settings or ask user
        const config = vscode.workspace.getConfiguration('typing-master');
        let githubToken = config.get<string>('githubToken');

        if (!githubToken) {
            githubToken = await vscode.window.showInputBox({
                prompt: 'Enter your GitHub personal access token',
                password: true,
                validateInput: (value) => {
                    if (!value || value.length < 10) {
                        return 'Please enter a valid GitHub token';
                    }
                    return null;
                }
            });

            if (!githubToken) {
                return;
            }

            // Ask if user wants to save the token
            const saveToken = await vscode.window.showQuickPick(['Yes', 'No'], {
                placeHolder: 'Save token to VS Code settings for future use?'
            });

            if (saveToken === 'Yes') {
                await config.update('githubToken', githubToken, vscode.ConfigurationTarget.Global);
            }
        }

        // Get repository info
        const repoOwner = await vscode.window.showInputBox({
            prompt: 'Enter repository owner (username or organization)',
            validateInput: (value) => {
                if (!value || value.trim().length === 0) {
                    return 'Repository owner is required';
                }
                return null;
            }
        });

        if (!repoOwner) return;

        const repoName = await vscode.window.showInputBox({
            prompt: 'Enter repository name',
            validateInput: (value) => {
                if (!value || value.trim().length === 0) {
                    return 'Repository name is required';
                }
                return null;
            }
        });

        if (!repoName) return;

        const filePath = await vscode.window.showInputBox({
            prompt: 'Enter file path in repository (e.g., snippets/example.js)',
            value: `snippets/${language}-snippet-${Date.now()}.${getFileExtension(language)}`,
            validateInput: (value) => {
                if (!value || value.trim().length === 0) {
                    return 'File path is required';
                }
                return null;
            }
        });

        if (!filePath) return;

        const commitMessage = await vscode.window.showInputBox({
            prompt: 'Enter commit message',
            value: `Add ${language} typing practice snippet`,
            validateInput: (value) => {
                if (!value || value.trim().length === 0) {
                    return 'Commit message is required';
                }
                return null;
            }
        });

        if (!commitMessage) return;

        // Ask if user wants to create a pull request
        const createPR = await vscode.window.showQuickPick(['Yes', 'No'], {
            placeHolder: 'Create a pull request?'
        });

        let prTitle: string | undefined;
        let prDescription: string | undefined;
        let branch: string | undefined;

        if (createPR === 'Yes') {
            prTitle = await vscode.window.showInputBox({
                prompt: 'Enter pull request title',
                value: `Add ${language} typing practice snippet`,
                validateInput: (value) => {
                    if (!value || value.trim().length === 0) {
                        return 'Pull request title is required';
                    }
                    return null;
                }
            });

            if (!prTitle) return;

            prDescription = await vscode.window.showInputBox({
                prompt: 'Enter pull request description (optional)',
                value: `This PR adds a new ${language} typing practice snippet generated from Typing Master for Coding.`
            });

            branch = await vscode.window.showInputBox({
                prompt: 'Enter branch name (optional)',
                value: `typing-master-snippet-${Date.now()}`
            });
        }

        // Show progress indicator
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Sharing to GitHub...',
            cancellable: false
        }, async (progress) => {
            progress.report({ increment: 0, message: 'Preparing to share...' });

            const response = await axios.post(`${apiUrl}/api/v1/integrations/github/share`, {
                access_token: githubToken,
                repo_owner: repoOwner.trim(),
                repo_name: repoName.trim(),
                file_path: filePath.trim(),
                content: sourceCode,
                commit_message: commitMessage.trim(),
                create_pr: createPR === 'Yes',
                pr_title: prTitle?.trim() || '',
                pr_description: prDescription?.trim() || '',
                branch: branch?.trim() || ''
            });

            progress.report({ increment: 100, message: 'Complete!' });

            const result = response.data;
            if (result.success) {
                const message = result.pull_request 
                    ? `Snippet shared successfully! Pull request created: ${result.pull_request.url}`
                    : `Snippet shared successfully! File URL: ${result.file_url}`;
                
                vscode.window.showInformationMessage(message, 'Open URL').then(selection => {
                    if (selection === 'Open URL') {
                        const url = result.pull_request?.url || result.file_url;
                        vscode.env.openExternal(vscode.Uri.parse(url));
                    }
                });
            } else {
                vscode.window.showErrorMessage(`Failed to share snippet: ${result.error}`);
            }
        });

    } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to share to GitHub: ${error.response?.data?.error || error.message}`);
    }
}

async function syncWithGitHub(context: vscode.ExtensionContext, apiUrl: string) {
    try {
        const config = vscode.workspace.getConfiguration('typing-master');
        let githubToken = config.get<string>('githubToken');

        if (!githubToken) {
            const result = await vscode.window.showQuickPick(['Enter Token', 'Cancel'], {
                placeHolder: 'GitHub token not found. Do you want to enter it?'
            });

            if (result === 'Enter Token') {
                await shareToGitHub(context, apiUrl); // This will prompt for token
                return;
            }
        }

        // Show progress indicator
        await vscode.window.withProgress({
            location: vscode.ProgressLocation.Notification,
            title: 'Syncing with GitHub...',
            cancellable: false
        }, async (progress) => {
            progress.report({ increment: 0, message: 'Fetching repositories...' });

            // Get user's repositories
            const response = await axios.get(`${apiUrl}/api/v1/integrations/github/repos`, {
                headers: {
                    'Authorization': `Bearer ${githubToken}`
                }
            });

            progress.report({ increment: 50, message: 'Processing repositories...' });

            const repos = response.data;
            if (repos.length === 0) {
                vscode.window.showInformationMessage('No repositories found with typing practice snippets.');
                return;
            }

            // Show repository selection
            const repoNames = repos.map((repo: any) => `${repo.full_name} (${repo.language || 'Unknown'})`);
            const selectedRepo = await vscode.window.showQuickPick(repoNames, {
                placeHolder: 'Select a repository to sync snippets from'
            });

            if (!selectedRepo) return;

            const repo = repos.find((r: any) => `${r.full_name} (${r.language || 'Unknown'})` === selectedRepo);
            
            progress.report({ increment: 100, message: 'Complete!' });

            vscode.window.showInformationMessage(
                `Synced with ${repo.full_name}. Found ${repo.snippet_count || 0} typing practice snippets.`,
                'View Snippets'
            ).then(selection => {
                if (selection === 'View Snippets') {
                    vscode.env.openExternal(vscode.Uri.parse(repo.html_url));
                }
            });
        });

    } catch (error: any) {
        vscode.window.showErrorMessage(`Failed to sync with GitHub: ${error.response?.data?.error || error.message}`);
    }
}

function getFileExtension(language: string): string {
    const extensions: { [key: string]: string } = {
        'javascript': 'js',
        'typescript': 'ts',
        'python': 'py',
        'java': 'java',
        'cpp': 'cpp',
        'rust': 'rs',
        'yaml': 'yaml',
        'json': 'json'
    };
    return extensions[language] || 'txt';
}
