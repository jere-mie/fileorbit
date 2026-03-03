/* ==========================================================================
   FileOrbit — Client-side Interactivity
   Handles: drag-drop uploads, search, copy-to-clipboard, toast notifications,
   auto-dismiss alerts, upload form state management.
   ========================================================================== */

(function () {
    'use strict';

    // ---- Toast Notifications ----
    function showToast(message) {
        const existing = document.querySelector('.toast');
        if (existing) existing.remove();

        const toast = document.createElement('div');
        toast.className = 'toast';
        toast.textContent = message;
        document.body.appendChild(toast);

        setTimeout(() => {
            if (toast.parentNode) toast.remove();
        }, 2500);
    }

    // ---- Copy to Clipboard ----
    window.copyToClipboard = function (text) {
        if (navigator.clipboard && navigator.clipboard.writeText) {
            navigator.clipboard.writeText(text).then(() => {
                showToast('Link copied to clipboard');
            }).catch(() => {
                fallbackCopy(text);
            });
        } else {
            fallbackCopy(text);
        }
    };

    function fallbackCopy(text) {
        const ta = document.createElement('textarea');
        ta.value = text;
        ta.style.position = 'fixed';
        ta.style.opacity = '0';
        document.body.appendChild(ta);
        ta.select();
        try {
            document.execCommand('copy');
            showToast('Link copied to clipboard');
        } catch (e) {
            showToast('Failed to copy');
        }
        document.body.removeChild(ta);
    }

    // ---- Auto-dismiss success alerts ----
    function initAlerts() {
        const successAlert = document.getElementById('success-alert');
        if (successAlert) {
            setTimeout(() => {
                successAlert.style.transition = 'opacity 0.4s ease, transform 0.4s ease';
                successAlert.style.opacity = '0';
                successAlert.style.transform = 'translateY(-8px)';
                setTimeout(() => successAlert.remove(), 400);
            }, 4000);
        }
    }

    // ---- Upload Zone: Drag & Drop + File Selection ----
    function initUpload() {
        const zone = document.getElementById('upload-zone');
        const input = document.getElementById('file-input');
        const options = document.getElementById('upload-options');
        const preview = document.getElementById('upload-file-preview');
        const cancelBtn = document.getElementById('upload-cancel');
        const form = document.getElementById('upload-form');
        const uploadBtn = document.getElementById('upload-btn');

        if (!zone || !input) return;

        // Click on zone opens file picker
        zone.addEventListener('click', function (e) {
            if (e.target.tagName !== 'LABEL' && e.target.tagName !== 'INPUT') {
                input.click();
            }
        });

        // Drag and drop
        zone.addEventListener('dragover', function (e) {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.add('drag-over');
        });

        zone.addEventListener('dragleave', function (e) {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.remove('drag-over');
        });

        zone.addEventListener('drop', function (e) {
            e.preventDefault();
            e.stopPropagation();
            zone.classList.remove('drag-over');

            if (e.dataTransfer.files.length > 0) {
                input.files = e.dataTransfer.files;
                showFileOptions(e.dataTransfer.files[0]);
            }
        });

        // File input change
        input.addEventListener('change', function () {
            if (input.files.length > 0) {
                showFileOptions(input.files[0]);
            }
        });

        function showFileOptions(file) {
            if (options) {
                options.classList.add('active');
            }
            if (preview) {
                preview.innerHTML = '';
                const badge = document.createElement('span');
                badge.className = 'file-type-badge';
                badge.textContent = getFileExt(file.name);
                const name = document.createElement('span');
                name.textContent = file.name;
                const size = document.createElement('span');
                size.style.color = 'var(--text-3)';
                size.style.marginLeft = 'auto';
                size.style.fontFamily = 'var(--font-mono)';
                size.style.fontSize = '0.75rem';
                size.textContent = formatFileSize(file.size);
                preview.appendChild(badge);
                preview.appendChild(name);
                preview.appendChild(size);
            }
        }

        // Cancel button
        if (cancelBtn) {
            cancelBtn.addEventListener('click', function () {
                if (options) options.classList.remove('active');
                input.value = '';
                if (preview) preview.innerHTML = '';
            });
        }

        // Loading state on submit
        if (form && uploadBtn) {
            form.addEventListener('submit', function () {
                const btnText = uploadBtn.querySelector('.btn-text');
                const btnLoading = uploadBtn.querySelector('.btn-loading');
                if (btnText) btnText.style.display = 'none';
                if (btnLoading) btnLoading.style.display = 'inline';
                uploadBtn.disabled = true;
                uploadBtn.style.opacity = '0.7';
            });
        }
    }

    // ---- Search ----
    function initSearch() {
        const searchInput = document.getElementById('search-input');
        const filesGrid = document.getElementById('files-grid');

        if (!searchInput || !filesGrid) return;

        let debounceTimer;
        const originalCards = filesGrid.innerHTML;

        searchInput.addEventListener('input', function () {
            clearTimeout(debounceTimer);
            const query = searchInput.value.trim().toLowerCase();

            if (query === '') {
                filesGrid.innerHTML = originalCards;
                return;
            }

            debounceTimer = setTimeout(function () {
                // Client-side filtering of existing cards
                const cards = filesGrid.querySelectorAll('.file-card');
                let anyVisible = false;

                cards.forEach(function (card) {
                    const filename = (card.dataset.filename || '').toLowerCase();
                    const description = (card.dataset.description || '').toLowerCase();
                    const url = (card.dataset.url || '').toLowerCase();

                    if (filename.includes(query) || description.includes(query) || url.includes(query)) {
                        card.style.display = '';
                        anyVisible = true;
                    } else {
                        card.style.display = 'none';
                    }
                });

                // Show empty state if nothing matches
                const existing = filesGrid.querySelector('.search-empty');
                if (!anyVisible && !existing) {
                    const empty = document.createElement('div');
                    empty.className = 'empty-state search-empty';
                    empty.innerHTML = '<p class="empty-text">No files match "' + escapeHtml(query) + '"</p>';
                    filesGrid.appendChild(empty);
                } else if (anyVisible && existing) {
                    existing.remove();
                }
            }, 150);
        });

        // Also handle server-side search for large datasets
        searchInput.addEventListener('keydown', function (e) {
            if (e.key === 'Enter') {
                e.preventDefault();
                const query = searchInput.value.trim();
                if (query.length === 0) return;

                fetch('/api/search?q=' + encodeURIComponent(query))
                    .then(function (res) { return res.json(); })
                    .then(function (data) {
                        if (data.files && data.files.length > 0) {
                            showToast(data.files.length + ' file(s) found');
                        }
                    })
                    .catch(function () {
                        // Silently fail, client-side search already works
                    });
            }
        });
    }

    // ---- Keyboard Shortcuts ----
    function initKeyboardShortcuts() {
        document.addEventListener('keydown', function (e) {
            // Ctrl/Cmd + K: Focus search
            if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
                e.preventDefault();
                const searchInput = document.getElementById('search-input');
                if (searchInput) searchInput.focus();
            }
        });
    }

    // ---- Helpers ----
    function getFileExt(name) {
        var dot = name.lastIndexOf('.');
        if (dot === -1) return 'FILE';
        return name.substring(dot + 1).toUpperCase();
    }

    function formatFileSize(bytes) {
        if (bytes < 1024) return bytes + ' B';
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
        if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
        return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
    }

    function escapeHtml(str) {
        var div = document.createElement('div');
        div.textContent = str;
        return div.innerHTML;
    }

    // ---- Initialize ----
    document.addEventListener('DOMContentLoaded', function () {
        initAlerts();
        initUpload();
        initSearch();
        initKeyboardShortcuts();
        initSingleTrigger();
    });

    // ---- "Single" Easter Egg: click 5 times to reveal login ----
    function initSingleTrigger() {
        var trigger = document.getElementById('single-trigger');
        var panel = document.getElementById('login-panel');
        if (!trigger || !panel) return;

        var clickCount = 0;
        var resetTimer = null;

        trigger.addEventListener('click', function () {
            clickCount++;

            // Visual pulse feedback
            trigger.classList.remove('pulse');
            void trigger.offsetWidth; // force reflow
            trigger.classList.add('pulse');

            // Reset counter if no click within 2 seconds
            clearTimeout(resetTimer);
            resetTimer = setTimeout(function () {
                clickCount = 0;
            }, 2000);

            if (clickCount >= 5) {
                clickCount = 0;
                clearTimeout(resetTimer);
                panel.classList.remove('hidden');
                panel.classList.add('reveal');
                // Focus the password input after animation
                setTimeout(function () {
                    var pw = document.getElementById('password');
                    if (pw) pw.focus();
                }, 350);
            }
        });
    }

    // ---- Base URL (derived from window.location.origin) ----
    document.querySelectorAll('.input-prefix').forEach(function (el) {
        el.textContent = window.location.origin + '/';
    });

    document.querySelectorAll('.file-url-text').forEach(function (el) {
        var card = el.closest('.file-card');
        if (card && card.dataset.url) {
            el.title = window.location.origin + '/' + card.dataset.url;
        }
    });
})();
