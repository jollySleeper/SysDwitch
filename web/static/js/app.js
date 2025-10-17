// Service Control Panel JavaScript
// Additional client-side functionality

// Add any custom JavaScript functionality here
console.log('Service Control Panel loaded');

// Example: Add loading states to buttons
function addLoadingState(button) {
    const originalText = button.textContent;
    button.textContent = 'Processing...';
    button.disabled = true;

    return () => {
        button.textContent = originalText;
        button.disabled = false;
    };
}

// Export for use in other scripts
window.ServiceControl = {
    addLoadingState
};
