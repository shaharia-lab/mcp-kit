import React, { useState } from 'react';
import { Message } from './Message';

interface ModelSettings {
    temperature: number;
    maxTokens: number;
    topP: number;
    topK: number;
}

interface ChatMessage {
    content: string;
    isUser: boolean;
}

interface ChatContainerProps {
    toolsEnabled: boolean;
    modelSettings: ModelSettings;
}

export const ChatContainer: React.FC<ChatContainerProps> = ({ toolsEnabled: initialToolsEnabled, modelSettings }) => {
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [inputValue, setInputValue] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [toolsEnabled, setToolsEnabled] = useState(initialToolsEnabled);


    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inputValue.trim() || isLoading) return;

        const newMessage: ChatMessage = {
            content: inputValue,
            isUser: true
        };

        setMessages(prev => [...prev, newMessage]);
        setInputValue('');
        setIsLoading(true);

        try {
            const response = await fetch('http://localhost:8081/ask', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    question: inputValue,
                    useTools: toolsEnabled,
                    modelSettings: {
                        temperature: modelSettings.temperature,
                        maxTokens: modelSettings.maxTokens,
                        topP: modelSettings.topP,
                        topK: modelSettings.topK
                    }
                }),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();

            // Updated validation to check for 'answer' instead of 'response'
            if (!data || typeof data.answer !== 'string') {
                throw new Error('Invalid response format from server');
            }

            setMessages(prev => [...prev, {
                content: data.answer, // Use data.answer instead of data.response
                isUser: false
            }]);
        } catch (error) {
            console.error('Error:', error);
            setMessages(prev => [...prev, {
                content: "Sorry, there was an error processing your request. Please try again.",
                isUser: false
            }]);
        } finally {
            setIsLoading(false);
        }


    };
    return (
        <div className="chat-container overflow-hidden flex flex-col">
            <div className="flex-1 overflow-y-auto p-4 space-y-4">
                {messages.map((msg, index) => (
                    <Message
                        key={index}
                        content={msg.content}
                        isUser={msg.isUser}
                    />
                ))}

                {isLoading && (
                    <div className="flex items-center justify-center p-4">
                        <div className="animate-pulse text-gray-500">
                            Processing...
                        </div>
                    </div>
                )}
            </div>

            <form onSubmit={handleSubmit} className="p-4 border-t">
                <div className="flex flex-col gap-2">
                    <div className="flex items-center gap-2 mb-2">
                        <button
                            type="button"
                            onClick={() => setToolsEnabled(!toolsEnabled)}
                            className={`px-3 py-1.5 text-sm rounded-lg transition-colors duration-200 ${
                                toolsEnabled
                                    ? 'bg-gray-800 text-white hover:bg-gray-700'
                                    : 'bg-gray-200 text-gray-800 hover:bg-gray-300'
                            }`}
                        >
                            {toolsEnabled ? 'Tools Enabled' : 'Tools Disabled'}
                        </button>
                    </div>
                    <div className="flex gap-2">
                        <input
                            type="text"
                            value={inputValue}
                            onChange={(e) => setInputValue(e.target.value)}
                            className="w-full p-2 border rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                            placeholder="Type your message..."
                            disabled={isLoading}
                        />
                        <button
                            type="submit"
                            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600 disabled:opacity-50"
                            disabled={isLoading || !inputValue.trim()}
                        >
                            Send
                        </button>
                    </div>
                </div>
            </form>
        </div>
    );
};
