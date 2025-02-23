import React from 'react';
import { WrenchScrewdriverIcon } from "@heroicons/react/16/solid";

interface ToolsToggleButtonProps {
    toolsEnabled: boolean;
    onToolsToggle: (enabled: boolean) => void;
}

export const ToolsToggle: React.FC<ToolsToggleButtonProps> = ({
                                                                        toolsEnabled,
                                                                        onToolsToggle
                                                                    }) => {
    return (
        <button
            type="button"
            onClick={() => onToolsToggle(!toolsEnabled)}
            className={`px-3 py-1.5 text-sm rounded-lg transition-colors duration-200 w-fit flex items-center gap-2 ${
                toolsEnabled
                    ? 'bg-gray-800 text-white hover:bg-gray-700'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
        >
            <WrenchScrewdriverIcon className="h-4 w-4" />
            Tools {toolsEnabled ? 'Enabled' : 'Disabled'}
        </button>
    );
};