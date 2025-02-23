import React, { useEffect, useState } from 'react';
import { ArrowPathIcon, QuestionMarkCircleIcon, Cog6ToothIcon } from '@heroicons/react/24/outline';
import {ChatHistory} from "../types";
import {fetchChatHistories} from "../api";
import {SidebarHeader} from "./sidebar/SidebarHeader.tsx";
import {NewChatSection} from "./sidebar/NewChatSection.tsx";

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
            } sidebar-width flex flex-col`}
        >
            <SidebarHeader onClose={onClose} />
            <NewChatSection onChatSelect={onChatSelect ?? (() => {})} />

            {/* Scrollable chat histories section */}
            <div className="flex-1 min-h-0 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-300 scrollbar-track-gray-100">
                <div className="p-4">
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
                            {chatHistories.map((chat) => (
                                <li
                                    key={chat.uuid}
                                    className={`p-2 rounded cursor-pointer ${
                                        selectedChatId === chat.uuid
                                            ? 'hover:bg-blue-100'
                                            : 'hover:bg-gray-50'
                                    }`}
                                    onClick={() => onChatSelect?.(chat.uuid)}
                                >
                                    <div className="text-sm font-medium">
                                        {getFirstMessage(chat)}
                                    </div>
                                    <div className="text-xs text-gray-500">
                                        {formatDate(chat.created_at)}
                                    </div>
                                </li>
                            ))}
                        </ul>
                    )}
                </div>
            </div>

            {/* Footer section - fixed */}
            <div className="p-4 border-t flex-shrink-0">
                <div className="flex space-x-4 justify-center">
                    <button className="p-2 hover:bg-gray-100 rounded" aria-label="Refresh">
                        <ArrowPathIcon className="h-5 w-5 text-gray-500" />
                    </button>
                    <button className="p-2 hover:bg-gray-100 rounded" aria-label="Help">
                        <QuestionMarkCircleIcon className="h-5 w-5 text-gray-500" />
                    </button>
                    <button className="p-2 hover:bg-gray-100 rounded" aria-label="Settings">
                        <Cog6ToothIcon className="h-5 w-5 text-gray-500" />
                    </button>
                </div>
            </div>
        </div>
    );
};