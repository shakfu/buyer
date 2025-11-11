// Toggle form visibility
function toggleForm(formId) {
    const form = document.getElementById(formId);
    if (form) {
        form.classList.toggle('hidden');
    }
}

// Auto-hide messages after 3 seconds
document.addEventListener('DOMContentLoaded', function() {
    const messages = document.querySelectorAll('.message');
    messages.forEach(function(message) {
        setTimeout(function() {
            message.style.opacity = '0';
            setTimeout(function() {
                message.remove();
            }, 300);
        }, 3000);
    });
});

// HTMX event handlers
document.body.addEventListener('htmx:afterSwap', function(event) {
    // Reset form after successful submission
    if (event.detail.successful && event.detail.target.tagName === 'FORM') {
        event.detail.target.reset();
    }
});

// Display error messages from server
document.body.addEventListener('htmx:responseError', function(event) {
    const errorText = event.detail.xhr.responseText || 'An error occurred';
    alert('Error: ' + errorText);
});
