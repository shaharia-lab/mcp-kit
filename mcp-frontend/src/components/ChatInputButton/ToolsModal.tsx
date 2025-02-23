import React, { useEffect, useState } from 'react';
import { XMarkIcon } from "@heroicons/react/24/outline";

interface Tool {
    name: string;
    description: string;
}

interface ToolsModalProps {
    isOpen: boolean;
    onClose: () => void;
    onSave: (selectedTools: string[]) => void;
    initialSelectedTools: string[];
}


export const ToolsModal: React.FC<ToolsModalProps> = ({
                                                          isOpen,
                                                          onClose,
                                                          onSave,
                                                          initialSelectedTools
                                                      }) => {
    const [tools, setTools] = useState<Tool[]>([]);
    const [selectedTools, setSelectedTools] = useState<string[]>(initialSelectedTools);


    useEffect(() => {
        const fetchTools = async () => {
            try {
                const response = await fetch('http://localhost:8081/api/tools');
                const data = await response.json();
                setTools(data);
            } catch (error) {
                console.error('Error fetching tools:', error);
            }
        };

        if (isOpen) {
            fetchTools();
        }
    }, [isOpen]);

    const handleToolToggle = (toolName: string) => {
        setSelectedTools(prev =>
            prev.includes(toolName)
                ? prev.filter(name => name !== toolName)
                : [...prev, toolName]
        );
    };

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl w-full max-w-md p-6">
                <div className="flex justify-between items-center mb-6">
                    <h2 className="text-xl font-semibold text-gray-800">Available Tools</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700 transition-colors"
                    >
                        <XMarkIcon className="h-6 w-6" />
                    </button>
                </div>

                <div className="space-y-4 max-h-[60vh] overflow-y-auto">
                    {tools.map((tool) => (
                        <div
                            key={tool.name}
                            className="flex items-start space-x-3 p-3 rounded-lg hover:bg-gray-50 transition-colors"
                        >
                            <input
                                type="checkbox"
                                id={tool.name}
                                checked={selectedTools.includes(tool.name)}
                                onChange={() => handleToolToggle(tool.name)}
                                className="mt-1 h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                            />
                            <label htmlFor={tool.name} className="flex-1 cursor-pointer">
                                <div className="font-medium text-gray-900">{tool.name}</div>
                                <div className="text-sm text-gray-500">{tool.description}</div>
                            </label>
                        </div>
                    ))}
                </div>

                <div className="mt-6 flex justify-end space-x-3">
                    <button
                        onClick={onClose}
                        className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => {
                            onSave(selectedTools);
                            onClose();
                        }}
                        className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 transition-colors"
                    >
                        Save Changes
                    </button>
                </div>
            </div>
        </div>
    );
};