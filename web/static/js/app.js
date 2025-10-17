// Service Control Panel JavaScript
// Dynamic service management functionality

// Load services from API
async function loadServices() {
    try {
        const response = await fetch('/api/services/status');
        const data = await response.json();
        if (data.services) {
            renderServices(data.services);
        }
    } catch (error) {
        console.error('Failed to load services:', error);
    }
}

// Render services in the grid
function renderServices(services) {
    const grid = document.getElementById('services-grid');
    grid.innerHTML = services.map(service => `
        <div class="bg-white rounded-lg shadow-md p-6 service-card">
            <div class="flex justify-between items-center mb-4">
                <h3 class="text-lg font-semibold">${service.name.replace('.service', '')}</h3>
                <span class="px-2 py-1 rounded-full text-sm ${service.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}">
                    ${service.status}
                </span>
            </div>
            <div class="flex gap-2">
                <button onclick="controlService('${service.name.replace('.service', '')}', 'start')"
                        class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded transition-colors ${service.active ? 'opacity-50 cursor-not-allowed' : ''}"
                        ${service.active ? 'disabled' : ''}>
                    Start
                </button>
                <button onclick="controlService('${service.name.replace('.service', '')}', 'stop')"
                        class="bg-red-500 hover:bg-red-600 text-white px-4 py-2 rounded transition-colors ${!service.active ? 'opacity-50 cursor-not-allowed' : ''}"
                        ${!service.active ? 'disabled' : ''}>
                    Stop
                </button>
            </div>
        </div>
    `).join('');
}

// Control service (start/stop)
async function controlService(serviceName, action) {
    try {
        const response = await fetch(`/api/services/${serviceName}/${action}`, {
            method: 'POST'
        });
        const result = await response.json();
        if (result.success) {
            loadServices(); // Refresh the display
        } else {
            alert('Operation failed: ' + (result.error || 'Unknown error'));
        }
    } catch (error) {
        console.error('Control service error:', error);
        alert('Operation failed');
    }
}

// Initialize when DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    console.log('Service Control Panel loaded');

    // Load services on page load
    loadServices();

    // Refresh every 30 seconds
    setInterval(loadServices, 30000);
});
