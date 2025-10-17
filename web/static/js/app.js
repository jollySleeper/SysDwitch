// Service Control Panel JavaScript
// Dynamic service management functionality

// Refresh services from API (for updates after actions)
async function refreshServices() {
    try {
        const response = await fetch('/api/services/status');
        const data = await response.json();
        if (data.services) {
            updateServiceCards(data.services);
        }
    } catch (error) {
        console.error('Failed to refresh services:', error);
    }
}

// Update service card states after actions
function updateServiceCards(services) {
    services.forEach(service => {
        const serviceName = service.name.replace('.service', '');
        const card = document.querySelector(`[data-service="${serviceName}"]`);
        if (card) {
            // Update status badge
            const statusBadge = card.querySelector('.status-badge');
            statusBadge.textContent = service.status;
            statusBadge.className = `px-2 py-1 rounded-full text-sm status-badge ${
                service.active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`;

            // Update buttons
            const startBtn = card.querySelector('.start-btn');
            const stopBtn = card.querySelector('.stop-btn');

            if (service.active) {
                startBtn.disabled = true;
                startBtn.classList.add('opacity-50', 'cursor-not-allowed');
                stopBtn.disabled = false;
                stopBtn.classList.remove('opacity-50', 'cursor-not-allowed');
            } else {
                startBtn.disabled = false;
                startBtn.classList.remove('opacity-50', 'cursor-not-allowed');
                stopBtn.disabled = true;
                stopBtn.classList.add('opacity-50', 'cursor-not-allowed');
            }
        }
    });
}

// Control service (start/stop)
async function controlService(serviceName, action) {
    try {
        const response = await fetch(`/api/services/${serviceName}/${action}`, {
            method: 'POST'
        });
        const result = await response.json();
        if (result.success) {
            refreshServices(); // Refresh the display after action
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

    // Add data-service attributes to cards for easier targeting
    document.querySelectorAll('.service-card').forEach((card, index) => {
        const serviceName = card.querySelector('h3').textContent;
        card.setAttribute('data-service', serviceName);

        // Add classes to buttons for easier targeting
        const buttons = card.querySelectorAll('button');
        if (buttons.length >= 2) {
            buttons[0].classList.add('start-btn');
            buttons[1].classList.add('stop-btn');
        }

        // Add class to status badge
        const statusBadge = card.querySelector('span');
        if (statusBadge) {
            statusBadge.classList.add('status-badge');
        }
    });

    // Refresh every 30 seconds to show status changes from external sources
    setInterval(refreshServices, 30000);
});
