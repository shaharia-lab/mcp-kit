import React, {useState, KeyboardEvent, useRef, useEffect} from 'react';
import { Message } from './Message';
import {WrenchScrewdriverIcon} from "@heroicons/react/16/solid";

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
    selectedChatId?: string; // Add this prop
}


export const ChatContainer: React.FC<ChatContainerProps> = ({
                                                                toolsEnabled: initialToolsEnabled,
                                                                modelSettings,
                                                                selectedChatId
                                                            }) => {

    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const [inputValue, setInputValue] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [toolsEnabled, setToolsEnabled] = useState(initialToolsEnabled);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const [chatUuid, setChatUuid] = useState<string | null>(null);


    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages, isLoading]);

    useEffect(() => {
        const loadChatHistory = async () => {
            if (!selectedChatId) {
                setMessages([]); // Clear messages for new chat
                setChatUuid(null);
                return;
            }

            try {
                const response = await fetch(`http://localhost:8081/chat/${selectedChatId}`);
                if (!response.ok) {
                    throw new Error('Failed to load chat history');
                }

                const data = await response.json();
                // Transform the messages to match our ChatMessage format
                const formattedMessages = data.messages.map((msg: any) => ({
                    content: msg.Text,
                    isUser: msg.IsUser
                }));

                setMessages(formattedMessages);
                setChatUuid(selectedChatId);
            } catch (error) {
                console.error('Error loading chat history:', error);
                // Optionally show an error message to the user
            }
        };

        loadChatHistory();
    }, [selectedChatId]);


    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit(e);
        }
    };

    const inputRef = useRef<HTMLTextAreaElement>(null);

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
            const payload = {
                question: inputValue,
                useTools: toolsEnabled,
                modelSettings,
                ...(chatUuid && { chat_uuid: chatUuid }) // Include chat_uuid if it exists
            };

            const response = await fetch('http://localhost:8081/ask', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(payload),
            });

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            if (!data || typeof data.answer !== 'string') {
                throw new Error('Invalid response format from server');
            }

            // Store the chat_uuid from the first response
            if (data.chat_uuid && !chatUuid) {
                setChatUuid(data.chat_uuid);
            }

            setMessages(prev => [...prev, {
                content: data.answer,
                isUser: false
            }]);

            inputRef.current?.focus();

        } catch (error) {
            console.error('Error:', error);
            setMessages(prev => [...prev, {
                content: "Sorry, there was an error processing your request. Please try again.",
                isUser: false
            }]);
        } finally {
            setIsLoading(false);
            inputRef.current?.focus();
        }
    };


    return (
        <div className="max-w-6xl mx-auto chat-container overflow-hidden flex flex-col h-full">
            <div className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
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
                <div ref={messagesEndRef} />
            </div>

            <form onSubmit={handleSubmit} className="p-4 border-t border-gray-200">
                <div className="flex flex-col gap-2">
                    <button
                        type="button"
                        onClick={() => setToolsEnabled(!toolsEnabled)}
                        className={`px-3 py-1.5 text-sm rounded-lg transition-colors duration-200 w-fit flex items-center gap-2 ${
                            toolsEnabled
                                ? 'bg-gray-800 text-white hover:bg-gray-700'
                                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                        }`}
                    >
                        <WrenchScrewdriverIcon className="h-4 w-4" />
                        Tools {toolsEnabled ? 'Enabled' : 'Disabled'}
                    </button>


                    <textarea
                        ref={inputRef}
                        value={inputValue}
                        onChange={(e) => setInputValue(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder="Type your message... (Press Enter to send, Shift+Enter for new line)"
                        className="w-full min-h-[80px] p-3 rounded-lg border border-gray-300 focus:border-gray-400 focus:ring-1 focus:ring-gray-400 outline-none resize-none"
                        disabled={isLoading}
                    />
                </div>
            </form>
        </div>
    );
};