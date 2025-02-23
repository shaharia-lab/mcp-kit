import React from 'react';
import {ContentParser} from "../../utils/contentParser.ts";

interface MessageContentProps {
    content: string;
}

export const MessageContent: React.FC<MessageContentProps> = ({ content }) => {
    const sanitizedContent = ContentParser.parse(content);

    return (
        <div
            className="prose"
            dangerouslySetInnerHTML={{ __html: sanitizedContent }}
        />
    );
};