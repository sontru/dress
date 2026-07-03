const manager = {
    type: 'photos',
    photos: [],
    categories: [],
    selected: new Set(),
    user: null,
    sort: 'newest',
    query: '',
};

const mediaConfig = {
    photos: {
        singular: 'photo',
        plural: 'photos',
        hubCategory: 'Photos',
        uploadTitle: 'Upload Photos',
        uploadHint: 'Add single photos or upload in bulk',
        dropTitle: 'Drag and drop photos here',
        dropSubtitle: 'PNG, JPEG or GIF up to 20MB each',
        accept: 'image/png,image/jpeg,image/gif',
        backed: true,
    },
    illustrations: {
        singular: 'illustration',
        plural: 'illustrations',
        hubCategory: 'Illustrations',
        uploadTitle: 'Upload Illustrations',
        uploadHint: 'Use the same workflow for illustrations',
        dropTitle: 'Drag and drop illustrations here',
        dropSubtitle: 'PNG, JPEG or GIF files',
        accept: 'image/png,image/jpeg,image/gif',
        backed: false,
    },
    vectors: {
        singular: 'vector',
        plural: 'vectors',
        hubCategory: 'Vectors',
        uploadTitle: 'Upload Vectors',
        uploadHint: 'Use the same workflow for vector packs',
        dropTitle: 'Drag and drop vectors here',
        dropSubtitle: 'SVG, EPS or AI files when vector storage is enabled',
        accept: '.svg,.eps,.ai',
        backed: false,
    },
    videos: {
        singular: 'video',
        plural: 'videos',
        hubCategory: 'Videos',
        uploadTitle: 'Upload Videos',
        uploadHint: 'Use the same workflow for clips and motion assets',
        dropTitle: 'Drag and drop videos here',
        dropSubtitle: 'MP4, MOV or WEBM files when video storage is enabled',
        accept: 'video/mp4,video/quicktime,video/webm',
        backed: false,
    },
};

const els = {};

document.addEventListener('DOMContentLoaded', () => {
    bindElements();
    bindEvents();
    initialiseManager();
});

function bindElements() {
    [
        'headerAvatar', 'headerName', 'profileAvatar', 'profileName', 'profileSince',
        'managerSearch', 'uploadForm', 'mediaInput', 'fileQueue', 'uploadTitle',
        'uploadHint', 'dropTitle', 'dropSubtitle', 'mediaTitle', 'mediaDescription',
        'mediaCategory', 'hubCategory', 'mediaTags', 'mediaPrice',
        'startUpload', 'uploadStatus', 'manageTitle', 'manageSummary', 'mediaTable',
        'selectedCount', 'selectAllButton', 'bulkEditButton', 'bulkCategorySelect',
        'bulkDeleteButton', 'sortSelect', 'categoryList', 'categorySummary',
        'createCategoryTop', 'uploadMediaTop', 'editDialog', 'editForm', 'editId',
        'editTitle', 'editDescription', 'editCategory', 'editHubCategory',
        'editPrice', 'editIsPublic', 'editStatus', 'categoryDialog', 'categoryForm',
        'newCategoryName', 'categoryStatus', 'storageUsed', 'mediaWatermark',
        'watermarkFileField', 'watermarkFile'
    ].forEach((id) => {
        els[id] = document.getElementById(id);
    });
}

function bindEvents() {
    document.querySelectorAll('.media-tab').forEach((tab) => {
        tab.addEventListener('click', () => setMediaType(tab.dataset.type));
    });

    els.managerSearch.addEventListener('input', debounce((event) => {
        manager.query = event.target.value.trim().toLowerCase();
        renderLibrary();
    }, 160));

    els.sortSelect.addEventListener('change', (event) => {
        manager.sort = event.target.value;
        renderLibrary();
    });

    els.mediaInput.addEventListener('change', renderFileQueue);
    els.mediaWatermark.addEventListener('change', updateWatermarkField);
    els.uploadForm.addEventListener('submit', uploadMedia);
    els.selectAllButton.addEventListener('click', selectAllVisible);
    els.bulkCategorySelect.addEventListener('change', moveSelectedToCategory);
    els.bulkDeleteButton.addEventListener('click', deleteSelected);
    els.bulkEditButton.addEventListener('click', editFirstSelected);
    els.createCategoryTop.addEventListener('click', openCategoryDialog);
    els.uploadMediaTop.addEventListener('click', () => {
        window.location.href = appPath('/upload');
    });
    els.categoryForm.addEventListener('submit', createCategory);
    els.editForm.addEventListener('submit', saveEdit);

    document.querySelectorAll('[data-close-dialog]').forEach((button) => {
        button.addEventListener('click', () => button.closest('dialog').close());
    });
}

async function initialiseManager() {
    await Promise.all([loadCurrentUser(), loadCategories()]);
    await loadPhotos();
    setMediaType('photos');
}

async function loadCurrentUser() {
    try {
        const response = await fetch(appPath('/api/me'));
        if (response.status === 401) {
            window.location.href = appPath('/login');
            return;
        }
        if (!response.ok) throw new Error(await response.text());
        manager.user = await response.json();
        renderUser();
    } catch (error) {
        console.error('Unable to load account', error);
    }
}

function renderUser() {
    const user = manager.user || {};
    const name = user.real_name || user.name || user.username || 'Your account';
    const since = user.member_since || 'recently';
    const avatar = user.avatar_url ? appPath(user.avatar_url) : avatarDataUri(name);

    if (els.headerAvatar) els.headerAvatar.src = avatar;
    els.profileAvatar.src = avatar;
    if (els.headerName) els.headerName.textContent = name;
    els.profileName.textContent = name;
    els.profileSince.textContent = `Member since ${since}`;
}

async function loadCategories() {
    try {
        const response = await fetch(appPath('/api/admin/categories'));
        if (response.status === 401) {
            window.location.href = appPath('/login');
            return;
        }
        if (!response.ok) throw new Error(await response.text());
        manager.categories = await response.json();
        renderCategoryControls();
    } catch (error) {
        console.error('Unable to load categories', error);
    }
}

async function loadPhotos() {
    try {
        const response = await fetch(appPath('/api/admin/photos'));
        if (response.status === 401) {
            window.location.href = appPath('/login');
            return;
        }
        if (!response.ok) throw new Error(await response.text());
        manager.photos = await response.json();
        manager.selected.clear();
        renderLibrary();
        renderCategories();
    } catch (error) {
        els.mediaTable.innerHTML = emptyState('Unable to load your media', 'Please refresh or sign in again.');
        console.error('Unable to load photos', error);
    }
}

function setMediaType(type) {
    manager.type = type;
    manager.selected.clear();
    const config = mediaConfig[type];

    document.querySelectorAll('.media-tab').forEach((tab) => {
        const active = tab.dataset.type === type;
        tab.classList.toggle('active', active);
        tab.setAttribute('aria-selected', String(active));
    });

    els.uploadTitle.textContent = config.uploadTitle;
    els.uploadHint.textContent = config.uploadHint;
    els.dropTitle.textContent = config.dropTitle;
    els.dropSubtitle.textContent = config.dropSubtitle;
    els.mediaInput.accept = config.accept;
    els.hubCategory.value = config.hubCategory;
    els.manageTitle.textContent = `Manage your ${config.plural}`;
    els.startUpload.textContent = `Start ${config.singular} upload`;
    els.uploadStatus.textContent = config.backed ? '' : `The ${config.plural} manager is designed and ready for its storage API.`;
    els.uploadStatus.className = 'status-line';
    updateWatermarkField();
    renderFileQueue();
    renderLibrary();
    renderCategories();
}

function getVisibleItems() {
    if (!mediaConfig[manager.type].backed) return [];

    const filtered = manager.photos.filter((photo) => {
        const haystack = [
            photo.title,
            photo.description,
            photo.category,
            photo.user_category,
            photo.price,
        ].join(' ').toLowerCase();
        return !manager.query || haystack.includes(manager.query);
    });

    return filtered.sort((a, b) => {
        if (manager.sort === 'title') return String(a.title).localeCompare(String(b.title));
        if (manager.sort === 'price') return Number(b.price || 0) - Number(a.price || 0);
        return new Date(b.created_at || 0) - new Date(a.created_at || 0);
    });
}

function renderLibrary() {
    const config = mediaConfig[manager.type];
    const items = getVisibleItems();
    els.manageSummary.textContent = config.backed
        ? `${items.length} ${config.plural} in this view`
        : `No ${config.plural} uploaded yet`;

    if (!config.backed) {
        els.mediaTable.innerHTML = emptyState(
            `${titleCase(config.plural)} workspace`,
            `This page uses the same upload, pricing, category and management layout for ${config.plural}.`
        );
        updateBulkBar();
        return;
    }

    if (!items.length) {
        els.mediaTable.innerHTML = emptyState('No photos found', 'Upload media or adjust your search.');
        updateBulkBar();
        return;
    }

    els.mediaTable.innerHTML = `
        <div class="media-table-head">
            <span></span>
            <span>Preview</span>
            <span>Title</span>
            <span>Category</span>
            <span>Price</span>
            <span>Status</span>
            <span>Uploaded</span>
            <span>Actions</span>
        </div>
        ${items.map(renderPhotoRow).join('')}
    `;

    els.mediaTable.querySelectorAll('[data-select-photo]').forEach((checkbox) => {
        checkbox.addEventListener('change', () => {
            const id = Number(checkbox.dataset.selectPhoto);
            if (checkbox.checked) {
                manager.selected.add(id);
            } else {
                manager.selected.delete(id);
            }
            updateBulkBar();
        });
    });

    els.mediaTable.querySelectorAll('[data-edit-photo]').forEach((button) => {
        button.addEventListener('click', () => openEditDialog(Number(button.dataset.editPhoto)));
    });

    els.mediaTable.querySelectorAll('[data-delete-photo]').forEach((button) => {
        button.addEventListener('click', () => deletePhoto(Number(button.dataset.deletePhoto)));
    });

    updateBulkBar();
}

function renderPhotoRow(photo) {
    const checked = manager.selected.has(photo.id) ? 'checked' : '';
    const statusClass = photo.is_public ? 'status-pill' : 'status-pill draft';
    const status = photo.is_public ? 'Published' : 'Draft';
    const uploaded = photo.created_at ? new Date(photo.created_at).toLocaleDateString(undefined, {
        month: 'short',
        day: 'numeric',
        year: 'numeric',
    }) : 'Unknown';

    const displayTitle = capitalizeFirstWord(photo.title || 'Untitled photo');

    return `
        <div class="media-row">
            <input type="checkbox" data-select-photo="${photo.id}" ${checked} aria-label="Select ${escapeHtml(displayTitle)}">
            <img class="media-preview" src="${escapeHtml(appPath(photo.thumbnail || photo.image_path || ''))}" alt="">
            <div class="media-title-cell">
                <strong>${escapeHtml(displayTitle)}</strong>
                <span>${escapeHtml(photo.file_type || 'Image')} ${escapeHtml(photo.dimensions || '')}</span>
            </div>
            <span class="pill">${escapeHtml(photo.user_category || 'Default')}</span>
            <span>$${formatPrice(photo.price)}</span>
            <span class="${statusClass}">${status}</span>
            <span>${uploaded}</span>
            <span class="row-actions">
                <button class="icon-button" type="button" data-edit-photo="${photo.id}" title="Edit" aria-label="Edit ${escapeHtml(displayTitle)}">
                    <svg viewBox="0 0 24 24"><path d="m4 20 4.5-1 10-10-3.5-3.5-10 10L4 20ZM14 7l3 3"/></svg>
                </button>
                <a class="icon-button" href="${appPath(`/photo/${photo.id}`)}" title="Open" aria-label="Open ${escapeHtml(displayTitle)}">
                    <svg viewBox="0 0 24 24"><path d="M7 17 17 7M9 7h8v8"/></svg>
                </a>
                <button class="icon-button" type="button" data-delete-photo="${photo.id}" title="Delete" aria-label="Delete ${escapeHtml(displayTitle)}">
                    <svg viewBox="0 0 24 24"><path d="M4 7h16M10 11v6M14 11v6M6 7l1 14h10l1-14M9 7V4h6v3"/></svg>
                </button>
            </span>
        </div>
    `;
}

function renderCategoryControls() {
    const options = manager.categories.length
        ? manager.categories.map((category) => `<option value="${escapeHtml(category.name)}">${escapeHtml(category.name)}</option>`).join('')
        : '<option value="Default">Default</option>';

    els.mediaCategory.innerHTML = options;
    els.editCategory.innerHTML = options;
    els.bulkCategorySelect.innerHTML = `<option value="">Move to category</option>${options}`;
}

function renderCategories() {
    const config = mediaConfig[manager.type];
    const counts = new Map();
    const previews = new Map();

    if (config.backed) {
        manager.photos.forEach((photo) => {
            const name = photo.user_category || 'Default';
            counts.set(name, (counts.get(name) || 0) + 1);
            if (!previews.has(name) && (photo.thumbnail || photo.image_path)) {
                previews.set(name, appPath(photo.thumbnail || photo.image_path));
            }
        });
    }

    els.categorySummary.textContent = `Manage your ${config.singular} categories`;
    els.categoryList.innerHTML = manager.categories.map((category) => {
        const count = counts.get(category.name) || 0;
        const preview = previews.get(category.name);
        const thumb = preview
            ? `<img src="${escapeHtml(preview)}" alt="">`
            : `<span class="category-thumb-fallback">${escapeHtml(category.name.slice(0, 1).toUpperCase())}</span>`;
        return `
            <article class="category-card">
                ${thumb}
                <div>
                    <strong>${escapeHtml(category.name)}</strong>
                    <span>${count} ${count === 1 ? config.singular : config.plural}</span>
                </div>
                <button class="icon-button" type="button" title="Category options" aria-label="Category options">
                    <svg viewBox="0 0 24 24"><path d="M12 5v.01M12 12v.01M12 19v.01"/></svg>
                </button>
            </article>
        `;
    }).join('') + `
        <button class="category-card" type="button" id="addCategoryCard">
            <span class="category-thumb-fallback">+</span>
            <div>
                <strong>Add new category</strong>
                <span>create group</span>
            </div>
            <span></span>
        </button>
    `;

    document.getElementById('addCategoryCard')?.addEventListener('click', openCategoryDialog);
}

function renderFileQueue() {
    const files = Array.from(els.mediaInput.files || []);
    if (files.length === 1) {
        els.mediaTitle.value = titleFromFilename(files[0].name);
    }
    els.fileQueue.hidden = files.length === 0;
    els.fileQueue.innerHTML = files.map((file) => `<span>${escapeHtml(file.name)} (${formatBytes(file.size)})</span>`).join('');
}

async function uploadMedia(event) {
    event.preventDefault();
    const config = mediaConfig[manager.type];
    if (!config.backed) {
        setStatus(els.uploadStatus, `${titleCase(config.plural)} need a backend upload endpoint before files can be saved.`, 'error');
        return;
    }

    const files = Array.from(els.mediaInput.files || []);
    if (!files.length) {
        setStatus(els.uploadStatus, 'Choose at least one file to upload.', 'error');
        return;
    }

    els.startUpload.disabled = true;
    setStatus(els.uploadStatus, `Uploading ${files.length} ${files.length === 1 ? 'file' : 'files'}...`);

    try {
        for (let index = 0; index < files.length; index += 1) {
            const form = new FormData();
            const file = files[index];
            const bulkMode = document.querySelector('input[name="uploadMode"]:checked')?.value === 'bulk';
            form.append('photo', file);
            form.append('title', bulkMode || !els.mediaTitle.value.trim() ? titleFromFilename(file.name) : els.mediaTitle.value.trim());
            form.append('description', els.mediaDescription.value.trim());
            form.append('user_category', els.mediaCategory.value || 'Default');
            form.append('category', config.hubCategory);
            form.append('tags', els.mediaTags.value.trim());
            form.append('price', els.mediaPrice.value || '29.99');
            form.append('photographer', manager.user?.name || manager.user?.real_name || manager.user?.username || '');
            if (els.mediaWatermark.checked) {
                form.append('apply_watermark', 'true');
                if (els.watermarkFile.files[0]) {
                    form.append('watermark', els.watermarkFile.files[0]);
                }
            }

            const response = await fetch(appPath('/api/admin/photos/upload'), {
                method: 'POST',
                body: form,
            });
            if (!response.ok) throw new Error(await response.text());
            setStatus(els.uploadStatus, `Uploaded ${index + 1} of ${files.length} files...`);
        }

        els.uploadForm.reset();
        els.mediaPrice.value = '29.99';
        updateWatermarkField();
        renderFileQueue();
        await Promise.all([loadCategories(), loadPhotos()]);
        setStatus(els.uploadStatus, 'Upload complete.', 'success');
    } catch (error) {
        setStatus(els.uploadStatus, error.message.trim() || 'Upload failed.', 'error');
    } finally {
        els.startUpload.disabled = false;
    }
}

function updateWatermarkField() {
    const enabled = Boolean(els.mediaWatermark?.checked);
    if (els.watermarkFileField) els.watermarkFileField.hidden = !enabled;
    if (!enabled && els.watermarkFile) els.watermarkFile.value = '';
}

function openEditDialog(photoId) {
    const photo = manager.photos.find((item) => item.id === photoId);
    if (!photo) return;

    els.editId.value = photo.id;
    els.editTitle.value = photo.title || '';
    els.editDescription.value = photo.description || '';
    els.editCategory.value = photo.user_category || 'Default';
    els.editHubCategory.value = photo.category || '';
    els.editPrice.value = formatPrice(photo.price);
    els.editIsPublic.checked = Boolean(photo.is_public);
    els.editStatus.textContent = '';
    els.editDialog.showModal();
}

async function saveEdit(event) {
    event.preventDefault();
    const id = Number(els.editId.value);
    const payload = {
        title: els.editTitle.value.trim(),
        description: els.editDescription.value.trim(),
        user_category: els.editCategory.value || 'Default',
        category: els.editHubCategory.value.trim(),
        price: Number(els.editPrice.value || 0),
        is_public: els.editIsPublic.checked,
    };

    try {
        const response = await fetch(appPath(`/api/photos/${id}`), {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload),
        });
        if (!response.ok) throw new Error(await response.text());
        await loadPhotos();
        setStatus(els.editStatus, 'Saved changes.', 'success');
        els.editDialog.close();
    } catch (error) {
        setStatus(els.editStatus, error.message.trim() || 'Unable to save changes.', 'error');
    }
}

async function deletePhoto(photoId) {
    if (!window.confirm('Delete this media item?')) return;

    try {
        const response = await fetch(appPath(`/api/photos/${photoId}`), { method: 'DELETE' });
        if (!response.ok) throw new Error(await response.text());
        manager.selected.delete(photoId);
        await loadPhotos();
    } catch (error) {
        window.alert(error.message.trim() || 'Unable to delete media.');
    }
}

function selectAllVisible() {
    getVisibleItems().forEach((photo) => manager.selected.add(photo.id));
    renderLibrary();
}

function updateBulkBar() {
    const count = manager.selected.size;
    els.selectedCount.textContent = String(count);
    els.bulkEditButton.disabled = count === 0;
    els.bulkDeleteButton.disabled = count === 0;
    els.bulkCategorySelect.disabled = count === 0;
    els.bulkCategorySelect.value = '';
}

function editFirstSelected() {
    const [first] = manager.selected;
    if (first) openEditDialog(first);
}

async function moveSelectedToCategory() {
    const category = els.bulkCategorySelect.value;
    if (!category) return;
    const selectedIds = Array.from(manager.selected);

    try {
        for (const id of selectedIds) {
            const photo = manager.photos.find((item) => item.id === id);
            if (!photo) continue;
            const payload = {
                title: photo.title,
                description: photo.description,
                category: photo.category,
                user_category: category,
                price: photo.price,
                is_public: photo.is_public,
            };
            const response = await fetch(appPath(`/api/photos/${id}`), {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });
            if (!response.ok) throw new Error(await response.text());
        }
        manager.selected.clear();
        await loadPhotos();
    } catch (error) {
        window.alert(error.message.trim() || 'Unable to move selected media.');
    }
}

async function deleteSelected() {
    const selectedIds = Array.from(manager.selected);
    if (!selectedIds.length || !window.confirm(`Delete ${selectedIds.length} selected item(s)?`)) return;

    try {
        for (const id of selectedIds) {
            const response = await fetch(appPath(`/api/photos/${id}`), { method: 'DELETE' });
            if (!response.ok) throw new Error(await response.text());
        }
        manager.selected.clear();
        await loadPhotos();
    } catch (error) {
        window.alert(error.message.trim() || 'Unable to delete selected media.');
    }
}

function openCategoryDialog() {
    els.newCategoryName.value = '';
    els.categoryStatus.textContent = '';
    els.categoryDialog.showModal();
    els.newCategoryName.focus();
}

async function createCategory(event) {
    event.preventDefault();
    const name = els.newCategoryName.value.trim();
    if (!name) return;

    try {
        const response = await fetch(appPath('/api/admin/categories'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name }),
        });
        if (!response.ok) throw new Error(await response.text());
        await loadCategories();
        renderCategories();
        setStatus(els.categoryStatus, 'Category created.', 'success');
        els.categoryDialog.close();
    } catch (error) {
        setStatus(els.categoryStatus, error.message.trim() || 'Unable to create category.', 'error');
    }
}

function emptyState(title, body) {
    return `
        <div class="empty-manager-state">
            <strong>${escapeHtml(title)}</strong>
            <span>${escapeHtml(body)}</span>
        </div>
    `;
}

function setStatus(element, message, type = '') {
    element.textContent = message;
    element.className = `status-line ${type}`.trim();
}

function formatPrice(value) {
    return Number(value || 0).toFixed(2);
}

function formatBytes(bytes) {
    if (!bytes) return '0 B';
    const units = ['B', 'KB', 'MB', 'GB'];
    let value = bytes;
    let unit = 0;
    while (value >= 1024 && unit < units.length - 1) {
        value /= 1024;
        unit += 1;
    }
    return `${value.toFixed(unit ? 1 : 0)} ${units[unit]}`;
}

function titleFromFilename(filename) {
    return filename
        .replace(/\.[^.]+$/, '')
        .replace(/[-_]+/g, ' ')
        .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function capitalizeFirstWord(value) {
    const title = String(value || '');
    return title.replace(/^(\s*)(\S)/, (_, leadingSpace, firstLetter) => leadingSpace + firstLetter.toUpperCase());
}

function titleCase(value) {
    return String(value).replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function escapeHtml(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;',
    }[char]));
}

function avatarDataUri(name) {
    const initial = encodeURIComponent(String(name || 'I').trim().slice(0, 1).toUpperCase() || 'I');
    return `data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='96' height='96' viewBox='0 0 96 96'%3E%3Crect width='96' height='96' rx='48' fill='%23dbeafe'/%3E%3Ctext x='48' y='57' text-anchor='middle' font-family='Arial' font-size='34' font-weight='700' fill='%231769ff'%3E${initial}%3C/text%3E%3C/svg%3E`;
}

function debounce(fn, wait) {
    let timeout;
    return (...args) => {
        window.clearTimeout(timeout);
        timeout = window.setTimeout(() => fn(...args), wait);
    };
}
