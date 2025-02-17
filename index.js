// Initialize marked
marked.setOptions({
    highlight: function (code, lang) {
        if (lang && hljs.getLanguage(lang)) {
            return hljs.highlight(code, {language: lang}).value;
        }
        return hljs.highlightAuto(code).value;
    },
    breaks: true
});

// DOM Elements
const sidebar = document.getElementById('sidebar');
const modelControls = document.getElementById('model-controls');
const sidebarToggle = document.getElementById('sidebar-toggle');
const chatControlToggle = document.getElementById('chat-control-toggle');
const closeControls = document.getElementById('close-controls');
const welcomeScreen = document.getElementById('welcome-screen');
const chatContainer = document.getElementById('chat-container');
const userInput = document.getElementById('user-input');
const sidebarOverlay = document.getElementById('sidebar-overlay');

// State
let sidebarOpen = false;
let modelControlsOpen = false;

// Toggle functions
function toggleSidebar() {
    sidebarOpen = !sidebarOpen;
    sidebar.style.transform = sidebarOpen ? 'translateX(0)' : 'translateX(-100%)';
    sidebarOverlay.classList.toggle('hidden');
    document.body.style.overflow = sidebarOpen ? 'hidden' : '';
}

function toggleModelControls() {
    modelControlsOpen = !modelControlsOpen;
    modelControls.style.transform = modelControlsOpen ? 'translateX(0)' : 'translateX(100%)';
}

// Event Listeners
sidebarToggle.addEventListener('click', toggleSidebar);
sidebarOverlay.addEventListener('click', toggleSidebar);
chatControlToggle.addEventListener('click', toggleModelControls);
closeControls.addEventListener('click', toggleModelControls);

// Message Functions
const createMessage = (content, isUser = false) => {
    const template = document.getElementById('message-template');
    const message = template.content.cloneNode(true);
    const messageDiv = message.querySelector('.message');
    const nameDiv = message.querySelector('.font-medium');
    const contentDiv = message.querySelector('.prose');
    const copyButton = message.querySelector('.copy-button');

    messageDiv.classList.add(
        isUser ? 'bg-gray-70' : 'bg-gray-100',
        'p-4',
        'rounded-lg'
    );
    messageDiv.setAttribute('data-message-user', isUser ? 'true' : 'false');
    messageDiv.setAttribute('data-message-assistant', isUser ? 'false' : 'true');
    nameDiv.textContent = isUser ? 'You' : 'Assistant';

    const sanitizedContent = DOMPurify.sanitize(marked.parse(content));
    contentDiv.innerHTML = sanitizedContent;

    copyButton.addEventListener('click', async () => {
        try {
            await navigator.clipboard.writeText(content);
            copyButton.textContent = 'Copied';
            copyButton.classList.add('text-green-500');
            setTimeout(() => {
                copyButton.textContent = '';
                copyButton.classList.remove('text-green-500');
            }, 1000);
        } catch (err) {
            console.error('Failed to copy text');
        }
    });

    return message;
};

// Mock API call
const getAIResponse = async (message) => {
    await new Promise(resolve => setTimeout(resolve, 1000));
    const mockResponses = [
        `# Heading 1\n\nThis is a test markdown content with a table:\n\n| Column 1 | Column 2 |\n|----------|----------|\n| Value 1  | Value 2  |\n\nAnd a code block:\n\n\`\`\`javascript\nconsole.log('Hello, World!');\n\`\`\`\n\nHere is an image:\n\n![Test Image](https://via.placeholder.com/150)\n\n## Subheading\n\n- Item 1\n- Item 2\n  - Subitem 2.1\n\nThank you!`,
        `## Another Response\n\n- **Bold Item**\n- *Italic Item*\n\n\`\`\`python\ndef add(a, b):\n    return a + b\n\`\`\`\n\nHere is another table:\n\n| Name       | Hobby        |\n|------------|--------------|\n| Alice      | Drawing      |\n| Bob        | Programming  |`,
        `### Response with Image\n\nHere is a test image:\n\n![Random Image](https://placehold.co/150x150/EEE/31343C)\n\nAlso, some inline code: \`let x = 10;\`. Below is markdown with a blockquote:\n\n> This is a blockquote example.\n\n### List Section\n\n1. Item A\n2. Item B\n   - Subitem B.1\n   - Subitem B.2`
    ];
    const randomIndex = Math.floor(Math.random() * mockResponses.length);
    return mockResponses[randomIndex];
};

// Send message handler
const sendMessage = async () => {
    const message = userInput.value.trim();
    if (!message) return;

    if (!chatContainer.querySelector('.message')) {
        welcomeScreen.classList.add('hidden');
        chatContainer.classList.remove('hidden');
    }

    try {
        chatContainer.appendChild(createMessage(message, true));
        userInput.value = '';
        userInput.style.height = 'auto';

        const response = await getAIResponse(message);
        chatContainer.appendChild(createMessage(response));

        chatContainer.scrollTop = chatContainer.scrollHeight;
    } catch (error) {
        console.error('Failed to get response:', error);
    }
};

// Keyboard event handler
userInput.addEventListener('keydown', async (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        await sendMessage();
    }
});

// Auto-resize textarea
userInput.addEventListener('input', () => {
    userInput.style.height = 'auto';
    userInput.style.height = Math.min(userInput.scrollHeight, 200) + 'px';
});

// Model settings handler
document.querySelectorAll('#model-controls input').forEach(input => {
    input.addEventListener('change', (e) => {
        console.log(`${e.target.previousElementSibling.textContent}: ${e.target.value}`);
    });
});

// Initialize
userInput.focus();