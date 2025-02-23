import React from 'react';
import { ClipboardDocumentIcon } from "@heroicons/react/16/solid";

interface MessageHeaderProps {
    isUser: boolean;
    onCopy: () => Promise<void>;
}

export const Header: React.FC<MessageHeaderProps> = ({ isUser, onCopy }) => {
    return (
        <div className="flex justify-between items-start mb-2">
            <div className="font-medium">
                {isUser ? 'You' : 'Assistant'}
            </div>
            {!isUser && (
                <button
                    onClick={onCopy}
                    className="copy-button text-gray-500 hover:text-gray-700"
                    title="Copy message"
                >
                    <ClipboardDocumentIcon className="h-5 w-5" />
                </button>
            )}
        </div>
    );
};