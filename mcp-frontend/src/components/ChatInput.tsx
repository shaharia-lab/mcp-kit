import React, { useState, KeyboardEvent } from 'react';
import { ToolsToggle } from './ChatInputButton/ToolsToggle.tsx';

interface ChatInputProps {
    onSubmit: (message: string) => Promise<void>;
    isLoading: boolean;
    toolsEnabled: boolean;
    onToolsToggle: (enabled: boolean) => void;
}

export const ChatInput: React.FC<ChatInputProps> = ({
                                                        onSubmit,
                                                        isLoading,
                                                        toolsEnabled,
                                                        onToolsToggle
                                                    }) => {
    const [inputValue, setInputValue] = useState('');

    const handleKeyDown = (e: KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit(e);
        }
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!inputValue.trim() || isLoading) return;

        await onSubmit(inputValue);
        setInputValue('');
    };

    return (
        <form onSubmit={handleSubmit} className="p-4 border-t border-gray-200">
            <div className="flex flex-col gap-2">
                <div className="flex gap-2">
                    <ToolsToggle
                        toolsEnabled={toolsEnabled}
                        onToolsToggle={onToolsToggle}
                    />
                </div>

                <textarea
                    value={inputValue}
                    onChange={(e) => setInputValue(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder="Type your message... (Press Enter to send, Shift+Enter for new line)"
                    className="w-full min-h-[80px] p-3 rounded-lg border border-gray-300 focus:border-gray-400 focus:ring-1 focus:ring-gray-400 outline-none resize-none"
                />
            </div>
        </form>
    );
};