import React from 'react';

interface SidebarProps {
    isOpen: boolean;
    onClose: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({ isOpen, onClose }) => {
    return (
        <div
            className={`fixed top-0 left-0 h-full bg-white shadow-lg z-40 transition-all duration-300 transform ${
                isOpen ? 'translate-x-0' : '-translate-x-full'
            } sidebar-width`}
        >
            <div className="p-4">
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-bold">Chat History</h2>
                    <button
                        onClick={onClose}
                        className="p-2 hover:bg-gray-100 rounded"
                        aria-label="Close sidebar"
                    >
                        ✕
                    </button>
                </div>

                {/* Chat History Section */}
                <div className="space-y-4">
                    <div className="border-b pb-4">
                        <h3 className="text-sm font-semibold text-gray-500 mb-2">Recent Chats</h3>
                        <ul className="space-y-2">
                            <li className="hover:bg-gray-50 p-2 rounded cursor-pointer">
                                <div className="text-sm font-medium">New Chat</div>
                                <div className="text-xs text-gray-500">Start a new conversation</div>
                            </li>
                        </ul>
                    </div>

                    {/* Saved Chats */}
                    <div>
                        <h3 className="text-sm font-semibold text-gray-500 mb-2">Saved Chats</h3>
                        <ul className="space-y-2">
                            {/* Placeholder for saved chats */}
                            <li className="text-sm text-gray-400 italic p-2">
                                No saved chats yet
                            </li>
                        </ul>
                    </div>
                </div>

                {/* Settings Section */}
                <div className="absolute bottom-0 left-0 right-0 p-4 border-t">
                    <div className="space-y-2">
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <span>🔄</span>
                            Clear History
                        </button>
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <span>❓</span>
                            Help & FAQ
                        </button>
                        <button
                            className="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 rounded flex items-center gap-2"
                        >
                            <span>⚙️</span>
                            Settings
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};