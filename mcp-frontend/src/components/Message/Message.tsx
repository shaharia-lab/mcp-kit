import React from 'react';
import { Header } from './Header.tsx';
import { Content } from './Content.tsx';

interface MessageProps {
    content: string;
    isUser: boolean;
}

export const Message: React.FC<MessageProps> = ({ content, isUser }) => {
    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(content);
            // Add copy feedback if needed
        } catch (err) {
            console.error('Failed to copy text:', err);
        }
    };

    return (
        <div
            className={`message p-4 rounded-lg ${
                isUser ? 'bg-lime-100' : 'bg-gray-100'
            }`}
            data-message-user={isUser}
            data-message-assistant={!isUser}
        >
            <Header
                isUser={isUser}
                onCopy={copyToClipboard}
            />
            <Content content={content} />
        </div>
    );
};