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
import { Cog6ToothIcon } from "@heroicons/react/24/outline";

interface ModelSettings {
    temperature: number;
    maxTokens: number;
    topP: number;
    topK: number;
}

function initializeMarked() {
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
    const [toolsEnabled] = useState(false);
    const [showWelcome, setShowWelcome] = useState(true);
    const [selectedChatId, setSelectedChatId] = useState<string | undefined>();
    const [modelSettings, setModelSettings] = useState<ModelSettings>({
        temperature: 0.5,
        maxTokens: 2000,
        topP: 0.5,
        topK: 50
    });

    useEffect(() => {
        initializeMarked();

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
            if (settings.temperature < 0 || settings.temperature > 2) {
                throw new Error('Temperature must be between 0 and 2');
            }
            if (settings.maxTokens < 50 || settings.maxTokens > 50000) {
                throw new Error('Max tokens must be between 50 and 50000');
            }
            if (settings.topP < 0 || settings.topP > 1) {
                throw new Error('TopP must be between 0 and 1');
            }
            if (settings.topK < 1 || settings.topK > 100) {
                throw new Error('TopK must be between 1 and 100');
            }

            setModelSettings(settings);
            setModelControlsOpen(false);
        } catch (error) {
            console.error('Invalid model settings:', error);
            alert(error instanceof Error ? error.message : 'Invalid model settings');
        }
    };

    const handleChatSelect = (chatId: string) => {
        setSelectedChatId(chatId);
        setShowWelcome(false); // Hide welcome screen when chat is selected
        toggleSidebar(); // Close sidebar after selection
    };

    return (
        <div className="bg-gray-50 min-h-screen">
            <div
                className={`fixed inset-0 bg-black bg-opacity-50 z-30 ${!sidebarOpen && 'hidden'}`}
                onClick={toggleSidebar}
            />

            <Sidebar
                isOpen={sidebarOpen}
                onClose={toggleSidebar}
                onChatSelect={handleChatSelect}
                selectedChatId={selectedChatId}
            />

            <ModelControls
                key={JSON.stringify(modelSettings)}
                isOpen={modelControlsOpen}
                onClose={toggleModelControls}
                onSave={handleModelControlsSave}
                initialSettings={modelSettings}
            />

            <div className="flex flex-col h-screen">
                <header className="bg-white shadow-sm w-full">
                    <div className="w-full px-4 sm:px-6 lg:px-8 py-4 flex justify-between items-center">
                        <button
                            onClick={toggleSidebar}
                            className="p-2 rounded-md hover:bg-gray-100 focus:outline-none"
                        >
                            <svg
                                className="h-6 w-6"
                                fill="none"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth="2"
                                viewBox="0 0 24 24"
                                stroke="currentColor"
                            >
                                <path d="M4 6h16M4 12h16M4 18h16"></path>
                            </svg>
                        </button>

                        <div className="flex items-center space-x-4">
                            <ToolsToggle />
                            <button
                                onClick={toggleModelControls}
                                className="p-2 rounded-md hover:bg-gray-100 focus:outline-none"
                                title="Model Settings"
                            >
                                <Cog6ToothIcon className="h-6 w-6" />
                            </button>
                        </div>
                    </div>
                </header>


                <main className="flex-1 overflow-hidden">
                    {showWelcome ? (
                        <WelcomeScreen onStart={() => setShowWelcome(false)} />

                    ) : (
                        <ChatContainer
                            toolsEnabled={toolsEnabled}
                            modelSettings={modelSettings}
                            selectedChatId={selectedChatId}
                        />
                    )}
                </main>
            </div>
        </div>
    );
}

export default App;