// Common utility functions shared across pages
(function() {
    'use strict';

    // Generic toggle edit function for table rows
    window.toggleEdit = function(id, prefix) {
        prefix = prefix || 'brand'; // Default to brand for backward compatibility
        const row = document.getElementById(prefix + '-' + id);
        const nameSpan = row.querySelector('.' + prefix + '-name, .spec-name, .product-name, .vendor-name');
        const form = row.querySelector('.edit-form');
        const editBtn = row.querySelector('.secondary');

        if (form.classList.contains('hidden')) {
            nameSpan.classList.add('hidden');
            form.classList.remove('hidden');
            editBtn.textContent = 'Save';
            editBtn.onclick = function() {
                htmx.trigger(form, 'submit');
            };
        } else {
            nameSpan.classList.remove('hidden');
            form.classList.add('hidden');
            editBtn.textContent = 'Edit';
            editBtn.onclick = function() { toggleEdit(id, prefix); };
        }
    };

    // Confirmation dialog for delete actions
    window.confirmDelete = function(message) {
        return confirm(message || 'Are you sure you want to delete this item?');
    };

})();
