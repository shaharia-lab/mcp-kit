import React, {useState, KeyboardEvent, useRef, useEffect} from 'react';
import { Message } from './Message';
import {WrenchScrewdriverIcon} from "@heroicons/react/16/solid";
import {chatService} from "../services/chatService.ts";
import {ChatInput} from "./ChatInput.tsx";

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
                const data = await chatService.loadChatHistory(selectedChatId);
                const formattedMessages = data.messages.map((msg: any) => ({
                    content: msg.Text,
                    isUser: msg.IsUser
                }));

                setMessages(formattedMessages);
                setChatUuid(selectedChatId);
            } catch (error) {
                console.error('Error loading chat history:', error);
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

    const handleMessageSubmit = async (message: string) => {
        setIsLoading(true);

        const newMessage: ChatMessage = {
            content: message,
            isUser: true
        };

        setMessages(prev => [...prev, newMessage]);

        try {
            const payload = {
                question: message,
                useTools: toolsEnabled,
                modelSettings,
                ...(chatUuid && { chat_uuid: chatUuid })
            };

            const data = await chatService.sendMessage(payload);

            if (data.chat_uuid && !chatUuid) {
                setChatUuid(data.chat_uuid);
            }

            setMessages(prev => [...prev, { content: data.answer, isUser: false }]);
        } catch (error) {
            console.error('Error:', error);
            setMessages(prev => [...prev, {
                content: "Sorry, there was an error processing your request.",
                isUser: false
            }]);
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="max-w-6xl mx-auto chat-container overflow-hidden flex flex-col h-full">
            <div className="flex-1 overflow-y-auto p-4 space-y-4 scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100 hover:scrollbar-thumb-gray-400">
                {messages.map((message, index) => (
                    <Message
                        key={index}
                        content={message.content}
                        isUser={message.isUser}
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
            <ChatInput
                onSubmit={handleMessageSubmit}
                isLoading={isLoading}
                toolsEnabled={toolsEnabled}
                onToolsToggle={setToolsEnabled}
            />
        </div>
    );
};
