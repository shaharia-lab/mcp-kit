import React, { useState, useEffect } from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import hljs from 'highlight.js';
import 'highlight.js/styles/github-dark.css';
import { Sidebar } from './components/Sidebar';
import { ModelControls } from './components/ModelControls';
import { ChatContainer } from './components/ChatContainer';
import { WelcomeScreen } from './components/WelcomeScreen';
import { ToolsToggle } from './components/ToolsToggle';

interface ModelSettings {
    temperature: number;
    maxTokens: number;
    frequencyPenalty: number;
    presencePenalty: number;
}

function initializeMarked() {
    // Using the correct types for marked options
    marked.setOptions({
        highlight: (code: string, language: string) => {
            if (language && hljs.getLanguage(language)) {
                try {
                    return hljs.highlight(code, { language }).value;
                } catch (e) {
                    console.error('Highlight.js error:', e);
                    return code;
                }
            }
            return hljs.highlightAuto(code).value;
        },
        breaks: true,
        gfm: true
    });

    // Configure DOMPurify
    DOMPurify.setConfig({
        ALLOWED_TAGS: [
            'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'p', 'br', 'hr',
            'code', 'pre', 'blockquote', 'strong', 'em', 'ul', 'ol',
            'li', 'a', 'table', 'thead', 'tbody', 'tr', 'th', 'td',
            'img', 'span', 'div'
        ],
        ALLOWED_ATTR: ['href', 'target', 'class', 'id', 'src', 'alt', 'title', 'style'],
        ALLOW_DATA_ATTR: false
    });
}

function App() {
    const [sidebarOpen, setSidebarOpen] = useState(false);
    const [modelControlsOpen, setModelControlsOpen] = useState(false);
    const [toolsEnabled, setToolsEnabled] = useState(false);
    const [showWelcome, setShowWelcome] = useState(true);
    const [modelSettings, setModelSettings] = useState<ModelSettings>({
        temperature: 0.7,
        maxTokens: 2000,
        frequencyPenalty: 0,
        presencePenalty: 0
    });

    useEffect(() => {
        initializeMarked();

        // Initialize highlight.js
        hljs.configure({
            ignoreUnescapedHTML: true,
            languages: [
                'javascript', 'typescript', 'python', 'java', 'go',
                'cpp', 'csharp', 'ruby', 'php', 'swift', 'kotlin',
                'rust', 'sql', 'bash', 'html', 'css', 'json', 'yaml',
                'markdown', 'xml'
            ]
        });
    }, []);

    const toggleSidebar = () => {
        setSidebarOpen(!sidebarOpen);
        document.body.style.overflow = !sidebarOpen ? 'hidden' : '';
    };

    const toggleModelControls = () => {
        setModelControlsOpen(!modelControlsOpen);
    };

    const handleModelControlsSave = (settings: ModelSettings) => {
        try {
            // Validate settings
            if (settings.temperature < 0 || settings.temperature > 1) {
                throw new Error('Temperature must be between 0 and 1');
            }
            if (settings.maxTokens < 1 || settings.maxTokens > 4000) {
                throw new Error('Max tokens must be between 1 and 4000');
            }
            if (settings.frequencyPenalty < -2 || settings.frequencyPenalty > 2) {
                throw new Error('Frequency penalty must be between -2 and 2');
            }
            if (settings.presencePenalty < -2 || settings.presencePenalty > 2) {
                throw new Error('Presence penalty must be between -2 and 2');
            }

            setModelSettings(settings);
            setModelControlsOpen(false);
        } catch (error) {
            console.error('Invalid model settings:', error);
            alert(error instanceof Error ? error.message : 'Invalid model settings');
        }
    };

    return (
        <div className="bg-gray-50 min-h-screen">
            {/* Overlay */}
            <div
                className={`fixed inset-0 bg-black bg-opacity-50 z-30 ${!sidebarOpen && 'hidden'}`}
                onClick={toggleSidebar}
            />

            <Sidebar isOpen={sidebarOpen} onClose={toggleSidebar} />
            <ModelControls
                isOpen={modelControlsOpen}
                onClose={toggleModelControls}
                onSave={handleModelControlsSave}
                initialSettings={modelSettings}
            />

            <main className="relative">
                <div className="flex items-center justify-between p-4">
                    <button
                        id="sidebar-toggle"
                        onClick={toggleSidebar}
                        className="p-2 hover:bg-gray-100 rounded"
                        aria-label="Toggle sidebar"
                    >
                        ☰
                    </button>

                    <div className="flex gap-2">
                        <ToolsToggle
                            enabled={toolsEnabled}
                            onChange={setToolsEnabled}
                        />
                        <button
                            id="chat-control-toggle"
                            onClick={toggleModelControls}
                            className="p-2 hover:bg-gray-100 rounded"
                            aria-label="Toggle model controls"
                        >
                            ⚙️
                        </button>
                    </div>
                </div>

                {showWelcome ? (
                    <WelcomeScreen onStart={() => setShowWelcome(false)} />
                ) : (
                    <ChatContainer
                        toolsEnabled={toolsEnabled}
                        modelSettings={modelSettings}
                    />
                )}
            </main>
        </div>
    );
}

export default App;