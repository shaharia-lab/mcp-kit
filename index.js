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
        isUser ? 'bg-lime-100' : 'bg-gray-100',
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

const getAIResponse = async (message) => {
    try {
        const response = await fetch('http://localhost:8081/ask', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                question: message
            })
        });

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        return data; // Return the whole response object
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
};


// Send message handler
const sendMessage = async () => {
    const message = userInput.value.trim();
    if (!message) return;

    if (!chatContainer.querySelector('.message')) {
        welcomeScreen.classList.add('hidden');
        chatContainer.classList.remove('hidden');
    }

    // Add user message
    chatContainer.appendChild(createMessage(message, true));
    userInput.value = '';
    userInput.style.height = 'auto';

    // Create loading message wrapper with a specific class
    const loadingWrapper = document.createElement('div');
    loadingWrapper.classList.add('loading-message');
    loadingWrapper.appendChild(createMessage("*Processing your request...*"));
    chatContainer.appendChild(loadingWrapper);
    chatContainer.scrollTop = chatContainer.scrollHeight;

    try {
        // Get AI response
        const response = await getAIResponse(message);

        // Remove loading message
        const loadingElement = document.querySelector('.loading-message');
        if (loadingElement) {
            loadingElement.remove();
        }

        // Check if response has the answer property and it's not empty
        if (response && typeof response.answer === 'string' && response.answer.trim()) {
            chatContainer.appendChild(createMessage(response.answer));
        } else {
            console.error('Invalid or empty response:', response);
            throw new Error("Invalid API response format");
        }
    } catch (error) {
        // Remove loading message
        const loadingElement = document.querySelector('.loading-message');
        if (loadingElement) {
            loadingElement.remove();
        }

        // Add error message
        const errorMessage = "Sorry, I couldn't process your request. Please try again.";
        chatContainer.appendChild(createMessage(errorMessage));
        console.error('Failed to get response:', error);
    } finally {
        chatContainer.scrollTop = chatContainer.scrollHeight;
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