import { marked } from 'marked';
import DOMPurify from 'dompurify';

export class ContentParser {
    static parse(rawContent: string): string {
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
    }
}