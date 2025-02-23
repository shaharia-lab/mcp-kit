// components/LLMProviderToggle.tsx
import React, { useState } from 'react';
import { CircleStackIcon } from "@heroicons/react/24/outline";
import { LLMProvidersModal } from './LLMProvidersModal';

interface LLMProviderToggleProps {
    selectedProvider: string | null;
    selectedModelId: string | null;
    onProviderChange: (provider: string, modelId: string) => void;
}

export const LLMProviderToggle: React.FC<LLMProviderToggleProps> = ({
                                                                        selectedProvider,
                                                                        selectedModelId,
                                                                        onProviderChange,
                                                                    }) => {
    const [isModalOpen, setIsModalOpen] = useState(false);

    const getDisplayText = () => {
        if (!selectedProvider) return 'Select LLM Provider';
        return `${selectedProvider} (${selectedModelId})`;
    };

    return (
        <>
            <button
                type="button"
                onClick={() => setIsModalOpen(true)}
                className={`px-3 py-1.5 text-sm rounded-lg transition-colors duration-200 w-fit flex items-center gap-2 ${
                    selectedProvider
                        ? 'bg-blue-600 text-white hover:bg-blue-700'
                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                }`}
            >
                <CircleStackIcon className="h-4 w-4" />
                {getDisplayText()}
            </button>

            <LLMProvidersModal
                isOpen={isModalOpen}
                onClose={() => setIsModalOpen(false)}
                onSave={onProviderChange}
                initialProvider={selectedProvider}
                initialModelId={selectedModelId}
            />
        </>
    );
};