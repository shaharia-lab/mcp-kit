import React from 'react';
import { marked } from 'marked';
import DOMPurify from 'dompurify';
import {ClipboardDocumentIcon} from "@heroicons/react/16/solid";

interface MessageProps {
    content: string;
    isUser: boolean;
}

export const Message: React.FC<MessageProps> = ({ content, isUser }) => {
    const parseAndSanitizeContent = (rawContent: string) => {
        try {
            if (!rawContent) {
                return 'No content available';
            }
            const parsedContent = marked(rawContent);
            return DOMPurify.sanitize(parsedContent);
        } catch (error) {
            console.error('Error parsing message content:', error);
            return 'Error displaying message content';
        }
    };

    const copyToClipboard = async () => {
        try {
            await navigator.clipboard.writeText(content);
            // Add copy feedback if needed
        } catch (err) {
            console.error('Failed to copy text:', err);
        }
    };

    const sanitizedContent = parseAndSanitizeContent(content);

    return (
        <div
            className={`message p-4 rounded-lg ${
                isUser ? 'bg-lime-100' : 'bg-gray-100'
            }`}
            data-message-user={isUser}
            data-message-assistant={!isUser}
        >
            <div className="flex justify-between items-start mb-2">
                <div className="font-medium">
                    {isUser ? 'You' : 'Assistant'}
                </div>
                {!isUser && (
                    <button
                        onClick={copyToClipboard}
                        className="copy-button text-gray-500 hover:text-gray-700"
                        title="Copy message"
                    >
                        <ClipboardDocumentIcon className="h-5 w-5" />
                    </button>

                )}
            </div>
            <div
                className="prose"
                dangerouslySetInnerHTML={{ __html: sanitizedContent }}
            />
        </div>
    );
};