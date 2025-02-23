import React, { useEffect, useState } from 'react';
import { ArrowPathIcon, QuestionMarkCircleIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';
import {ChatHistory} from "../types";
import {fetchChatHistories} from "../api";

interface SidebarProps {
    isOpen: boolean;
    onClose: () => void;
    onChatSelect?: (chatId: string) => void;
    selectedChatId?: string;
}

export const Sidebar: React.FC<SidebarProps> = ({
                                                    isOpen,
                                                    onClose,
                                                    onChatSelect,
                                                    selectedChatId
                                                }) => {
    const [chatHistories, setChatHistories] = useState<ChatHistory[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        const loadChatHistories = async () => {
            setIsLoading(true);
            try {
                const histories = await fetchChatHistories();
                setChatHistories(histories);
                setError(null);
            } catch (err) {
                setError('Failed to load chat histories');
                console.error(err);
            } finally {
                setIsLoading(false);
            }
        };

        loadChatHistories();
    }, []);

    const getFirstMessage = (chat: ChatHistory): string => {
        if (!chat.messages || chat.messages.length === 0) {
            return 'Untitled Chat';
        }

        const firstMessage = chat.messages[0].Text?.trim() || '';
        if (!firstMessage) {
            return 'Untitled Chat';
        }

        return firstMessage.length > 30
            ? `${firstMessage.substring(0, 30)}...`
            : firstMessage;
    };

    const formatDate = (dateString: string): string => {
        return new Date(dateString).toLocaleDateString();
    };

    return (
        <div
            className={`fixed top-0 left-0 h-full bg-white shadow-lg z-40 transition-all duration-300 transform ${
                isOpen ? 'translate-x-0' : '-translate-x-full'
            } sidebar-width`}
        >
            <div className="p-4">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-bold">Chat History</h2>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-100 rounded"
                        aria-label="Close sidebar"
                    >
                        ✕
                    </button>
                </div>

                {/* Chat History Section */}
                <div className="space-y-4">
                    <div className="border-b pb-4">
                        <h3 className="text-sm font-semibold text-gray-500 mb-2">Recent Chats</h3>
                        <ul className="space-y-2">
                            <li
                                className="hover:bg-gray-50 p-2 rounded cursor-pointer"
                                onClick={() => onChatSelect?.('')}
                            >
                                <div className="text-sm font-medium">New Chat</div>
                                <div className="text-xs text-gray-500">Start a new conversation</div>
                            </li>
                        </ul>
                    </div>

                    {/* Saved Chats */}
                    <div>
                        <h3 className="text-sm font-semibold text-gray-500 mb-2">Saved Chats</h3>
                        {isLoading ? (
                            <div className="text-sm text-gray-400 italic p-2">
                                Loading...
                            </div>
                        ) : error ? (
                            <div className="text-sm text-red-500 italic p-2">
                                {error}
                            </div>
                        ) : (
                            <ul className="space-y-2">
                                {chatHistories.length === 0 ? (
                                    <li className="text-sm text-gray-400 italic p-2">
                                        No saved chats yet
                                    </li>
                                ) : (
                                    chatHistories.map((chat) => {
                                        const messagePreview = getFirstMessage(chat);
                                        const messageDate = formatDate(chat.created_at);
                                        const messageCount = chat.messages?.length ?? 0;

                                        return (
                                            <li
                                                key={chat.uuid}
                                                onClick={() => onChatSelect?.(chat.uuid)}
                                                className={`p-2 rounded cursor-pointer transition-colors ${
                                                    selectedChatId === chat.uuid
                                                        ? 'bg-blue-50 hover:bg-blue-100'
                                                        : 'hover:bg-gray-50'
                                                }`}
                                            >
                                                <div className="text-sm font-medium truncate">
                                                    {messagePreview}
                                                </div>
                                                <div className="text-xs text-gray-500 flex justify-between items-center mt-1">
                                                    <span>{messageDate}</span>
                                                    <span>{messageCount} messages</span>
                                                </div>
                                            </li>
                                        );
                                    })
                                )}
                            </ul>
                        )}
                    </div>
                </div>

                {/* Settings Section */}
                <div className="absolute bottom-0 left-0 right-0 p-4 border-t">
                    <div className="space-y-2">
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <ArrowPathIcon className="h-5 w-5 text-gray-500" />
                            Clear History
                        </button>
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <QuestionMarkCircleIcon className="h-5 w-5 text-gray-500" />
                            Help & FAQ
                        </button>
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <Cog6ToothIcon className="h-5 w-5 text-gray-500" />
                            Settings
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};