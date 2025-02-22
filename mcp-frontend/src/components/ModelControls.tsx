import React, { useState } from 'react';

interface ModelControlsProps {
    isOpen: boolean;
    onClose: () => void;
}

export const ModelControls: React.FC<ModelControlsProps> = ({ isOpen, onClose }) => {
    const [temperature, setTemperature] = useState(0.7);
    const [maxTokens, setMaxTokens] = useState(2000);
    const [frequencyPenalty, setFrequencyPenalty] = useState(0);
    const [presencePenalty, setPresencePenalty] = useState(0);

    return (
        <div
            className={`fixed top-0 right-0 h-full w-[260px] bg-white shadow-lg z-40 transition-transform duration-300 ${
                isOpen ? 'translate-x-0' : 'translate-x-full'
            }`}
        >
            <div className="p-4">
                <button
                    onClick={onClose}
                    className="float-right p-2 hover:bg-gray-100 rounded"
                >
                    ✕
                </button>
                <h2 className="text-lg font-bold mb-4">Model Controls</h2>

                {/* Temperature Control */}
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Temperature: {temperature}
                    </label>
                    <input
                        type="range"
                        min="0"
                        max="2"
                        step="0.1"
                        value={temperature}
                        onChange={(e) => setTemperature(parseFloat(e.target.value))}
                        className="w-full"
                    />
                    <div className="flex justify-between text-xs text-gray-500">
                        <span>Precise</span>
                        <span>Balanced</span>
                        <span>Creative</span>
                    </div>
                </div>

                {/* Max Tokens Control */}
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Max Tokens: {maxTokens}
                    </label>
                    <input
                        type="range"
                        min="100"
                        max="4000"
                        step="100"
                        value={maxTokens}
                        onChange={(e) => setMaxTokens(parseInt(e.target.value))}
                        className="w-full"
                    />
                </div>

                {/* Frequency Penalty Control */}
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Frequency Penalty: {frequencyPenalty}
                    </label>
                    <input
                        type="range"
                        min="-2"
                        max="2"
                        step="0.1"
                        value={frequencyPenalty}
                        onChange={(e) => setFrequencyPenalty(parseFloat(e.target.value))}
                        className="w-full"
                    />
                </div>

                {/* Presence Penalty Control */}
                <div className="mb-4">
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                        Presence Penalty: {presencePenalty}
                    </label>
                    <input
                        type="range"
                        min="-2"
                        max="2"
                        step="0.1"
                        value={presencePenalty}
                        onChange={(e) => setPresencePenalty(parseFloat(e.target.value))}
                        className="w-full"
                    />
                </div>

                {/* Save Button */}
                <button
                    className="w-full bg-blue-600 text-white py-2 px-4 rounded hover:bg-blue-700 transition-colors"
                >
                    Save Changes
                </button>
            </div>
        </div>
    );
};