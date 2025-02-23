interface ChatResponse {
    answer: string;
    chat_uuid?: string;
}

interface ChatPayload {
    question: string;
    useTools: boolean;
    modelSettings: {
        temperature: number;
        maxTokens: number;
        topP: number;
        topK: number;
    };
    chat_uuid?: string;
}

export const chatService = {
    async sendMessage(payload: ChatPayload): Promise<ChatResponse> {
        const response = await fetch('http://localhost:8081/ask', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        return response.json();
    },

    async loadChatHistory(chatId: string) {
        const response = await fetch(`http://localhost:8081/chat/${chatId}`);
        if (!response.ok) {
            throw new Error('Failed to load chat history');
        }
        return response.json();
    }
};