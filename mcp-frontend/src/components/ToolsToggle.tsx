import React from 'react';

interface ToolsToggleProps {
    enabled: boolean;
    onChange: (enabled: boolean) => void;
}

export const ToolsToggle: React.FC<ToolsToggleProps> = ({ enabled, onChange }) => {
    return (
        <button
            onClick={() => onChange(!enabled)}
            className={`p-2 rounded transition-colors ${
                enabled
                    ? 'bg-gray-800 text-white hover:bg-white hover:text-gray-800'
                    : 'hover:bg-gray-100'
            }`}
        >
            🛠️
        </button>
    );
};