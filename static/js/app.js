// App state
const app = {
    currentPage: 1,
    pageSize: 'all',
    mediaCategory: document.body.dataset.mediaCategory || '',
    mediaLabel: document.body.dataset.mediaLabel || 'media',
    photos: [],
    totalPhotos: 0,
    selectedPhoto: null,
    collections: [],
    filters: {
        category: null,
        tags: [],
        searchQuery: ''
    }
};

// DOM Elements
const photosGrid = document.getElementById('photosGrid');
const searchInput = document.getElementById('searchInput');
const photoModal = document.getElementById('photoModal');
const collectionModal = document.getElementById('collectionModal');
const categoriesList = document.getElementById('categoriesList');
const tagsList = document.getElementById('tagsList');
const collectionsList = document.getElementById('collectionsList');

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadPhotos();
    loadCategories();
    loadTags();
    loadCollections();
    setupEventListeners();
    setupAuthButton();
});

// Setup event listeners
function setupEventListeners() {
    // Search
    searchInput.addEventListener('input', debounce(() => {
        app.currentPage = 1;
        loadPhotos();
    }, 300));

    // Pagination
    document.getElementById('prevBtn').addEventListener('click', () => {
        if (app.currentPage > 1) {
            app.currentPage--;
            loadPhotos();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    });

    document.getElementById('nextBtn').addEventListener('click', () => {
        const maxPages = Math.ceil(app.totalPhotos / app.pageSize);
        if (app.currentPage < maxPages) {
            app.currentPage++;
            loadPhotos();
            window.scrollTo({ top: 0, behavior: 'smooth' });
        }
    });

    // Collections
    document.getElementById('newCollectionBtn').addEventListener('click', () => {
        collectionModal.classList.add('active');
    });

    document.getElementById('collectionForm').addEventListener('submit', (e) => {
        e.preventDefault();
        createCollection();
    });

    // Modal close buttons
    document.querySelectorAll('.modal-close').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.target.closest('.modal').classList.remove('active');
        });
    });

    // Modal background click
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
    });

    const likeButton = document.getElementById('likePhotoBtn');
    if (likeButton) {
        likeButton.addEventListener('click', () => {
            if (app.selectedPhoto) toggleLike(app.selectedPhoto.id);
        });
    }
}

async function setupAuthButton() {
    const authButton = document.getElementById('authButton');
    const myMediaButton = document.getElementById('myMediaButton');
    const registerButton = document.getElementById('registerButton');
    if (!authButton) return;

    try {
        const response = await fetch(appPath('/api/me'));
        if (!response.ok) {
            authButton.textContent = 'Sign in';
            authButton.href = appPath('/login');
            if (myMediaButton) myMediaButton.style.display = 'none';
            if (registerButton) registerButton.style.display = '';
            return;
        }

        authButton.textContent = 'Sign out';
        authButton.href = '#';
        if (myMediaButton) myMediaButton.style.display = '';
        if (registerButton) registerButton.style.display = 'none';
        authButton.addEventListener('click', async (event) => {
            event.preventDefault();
            authButton.style.pointerEvents = 'none';

            try {
                const logoutResponse = await fetch(appPath('/api/logout'), { method: 'POST' });
                if (!logoutResponse.ok) throw new Error(await logoutResponse.text());
                appNavigate('/');
            } catch (error) {
                console.error('Error signing out:', error);
                showAlert('Error signing out', 'error');
                authButton.style.pointerEvents = '';
            }
        });
    } catch (error) {
        authButton.textContent = 'Sign in';
        authButton.href = appPath('/login');
        if (myMediaButton) myMediaButton.style.display = 'none';
        if (registerButton) registerButton.style.display = '';
    }
}

// Load photos
async function loadPhotos() {
    try {
        let url = `/api/photos?page=${app.currentPage}&pageSize=${app.pageSize}`;
        if (app.mediaCategory) url += `&category=${encodeURIComponent(app.mediaCategory)}`;

        // If search or filters applied, use search endpoint
        if (app.filters.searchQuery || app.filters.tags.length > 0) {
            url = `/api/search?q=${encodeURIComponent(app.filters.searchQuery)}`;
            if (app.mediaCategory) {
                url += `&category=${encodeURIComponent(app.mediaCategory)}`;
            } else if (app.filters.category) {
                url += `&category=${encodeURIComponent(app.filters.category)}`;
            }
            if (app.filters.tags.length > 0) url += `&tags=${app.filters.tags.join(',')}`;
            url += `&page=${app.currentPage}`;
        }

        const response = await fetch(appPath(url));
        const data = await response.json();

        app.photos = data.photos || [];
        app.totalPhotos = data.total || 0;

        renderPhotos();
        updatePaginationUI();
    } catch (error) {
        console.error('Error loading images:', error);
        showAlert('Error loading images', 'error');
    }
}

// Render images grid
function renderPhotos() {
    if (app.photos.length === 0) {
        photosGrid.innerHTML = `
            <div class="empty-state" style="grid-column: 1 / -1;">
                <div class="empty-state-icon">📷</div>
                <h3>No ${escapeHtml(app.mediaLabel)} found</h3>
                <p>Try adjusting your search or filters</p>
            </div>
        `;
        return;
    }

    photosGrid.innerHTML = app.photos.map(photo => `
        <div class="photo-card" onclick="openPhotoDetail(${photo.id})">
            <img src="${appPath(photo.thumbnail || photo.image_path)}" alt="${capitalizeFirstWord(photo.title)}" class="photo-card-image" />
            <div class="photo-card-info">
                <div class="photo-card-title">${capitalizeFirstWord(photo.title)}</div>
                ${photoCardMetadata(photo)}
                <button class="photo-card-likes ${photo.liked_by_user ? 'liked' : ''}" onclick="event.stopPropagation(); toggleLike(${photo.id})" aria-label="${photo.liked_by_user ? 'Unlike' : 'Like'} ${capitalizeFirstWord(photo.title)}">
                    <span>${photo.liked_by_user ? '♥' : '♡'}</span>
                    <span>${photo.like_count || 0}</span>
                </button>
            </div>
        </div>
    `).join('');
}

function photoCardMetadata(photo) {
    if (app.mediaCategory !== 'Photos') return '';
    const details = [
        photo.captured_at,
        photo.photo_location,
        photo.focal_length,
    ].filter(Boolean);
    if (!details.length) return '';
    return `<div class="photo-card-meta">${details.map(escapeHtml).join(' · ')}</div>`;
}

// Update pagination UI
function updatePaginationUI() {
    if (app.pageSize === 'all') {
        document.getElementById('pageInfo').textContent = `All images (${app.totalPhotos})`;
        document.getElementById('prevBtn').disabled = true;
        document.getElementById('nextBtn').disabled = true;
        return;
    }

    const maxPages = Math.ceil(app.totalPhotos / app.pageSize);
    document.getElementById('pageInfo').textContent = `Page ${app.currentPage} of ${maxPages || 1}`;
    document.getElementById('prevBtn').disabled = app.currentPage <= 1;
    document.getElementById('nextBtn').disabled = app.currentPage >= maxPages;
}

// Load categories
async function loadCategories() {
    try {
        const response = await fetch(appPath('/api/categories'));
        const categories = await response.json();

        categoriesList.innerHTML = categories.map(cat => `
            <div class="filter-item" onclick="filterByCategory('${cat.name}')">
                <input type="checkbox" id="cat-${cat.id}" />
                <label for="cat-${cat.id}">${cat.name}</label>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading categories:', error);
    }
}

// Load tags
async function loadTags() {
    try {
        const response = await fetch(appPath('/api/tags'));
        const tags = await response.json();

        tagsList.innerHTML = tags.map(tag => `
            <div class="filter-item" onclick="filterByTag('${tag.name}')">
                <input type="checkbox" id="tag-${tag.id}" />
                <label for="tag-${tag.id}">${tag.name}</label>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading tags:', error);
    }
}

// Filter by category
function filterByCategory(categoryName) {
    if (app.filters.category === categoryName) {
        app.filters.category = null;
    } else {
        app.filters.category = categoryName;
    }
    app.currentPage = 1;
    loadPhotos();
}

// Filter by tag
function filterByTag(tagName) {
    const index = app.filters.tags.indexOf(tagName);
    if (index > -1) {
        app.filters.tags.splice(index, 1);
    } else {
        app.filters.tags.push(tagName);
    }
    app.currentPage = 1;
    loadPhotos();
}

// Search
searchInput.addEventListener('change', (e) => {
    app.filters.searchQuery = e.target.value;
    app.currentPage = 1;
    loadPhotos();
});

// Navigate to photo detail page
function openPhotoDetail(photoId) {
    appNavigate(`/photo/${photoId}`);
}

// Load collections
async function loadCollections() {
    try {
        const response = await fetch(appPath('/api/collections?user_id=1'));
        const collections = await response.json();
        app.collections = collections;

        collectionsList.innerHTML = collections.map(col => `
            <div class="collection-item" onclick="viewCollection(${col.id})">
                <h4>${col.name}</h4>
                <span>${col.photo_count} images</span>
            </div>
        `).join('');
    } catch (error) {
        console.error('Error loading collections:', error);
    }
}

// Create collection
async function createCollection() {
    const name = document.getElementById('collectionName').value;
    const description = document.getElementById('collectionDesc').value;

    try {
        const response = await fetch(appPath('/api/collections'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                name,
                description,
                user_id: 1
            })
        });

        if (response.ok) {
            document.getElementById('collectionForm').reset();
            collectionModal.classList.remove('active');
            loadCollections();
            showAlert('Collection created successfully', 'success');
        }
    } catch (error) {
        console.error('Error creating collection:', error);
        showAlert('Error creating collection', 'error');
    }
}

// View collection
function viewCollection(collectionId) {
    // Could open a modal to show collection details
    console.log('View collection:', collectionId);
}

// Add to collection
function showAddToCollection() {
    if (app.collections.length === 0) {
        showAlert('Create a collection first', 'info');
        return;
    }

    const html = `
        <div class="modal active" id="addToCollectionModal">
            <div class="modal-content modal-sm">
                <button class="modal-close">&times;</button>
                <h3>Add to Collection</h3>
                <div style="display: flex; flex-direction: column; gap: 10px;">
                    ${app.collections.map(col => `
                        <button onclick="addPhotoToCollection(${col.id})" class="btn-primary" style="text-align: left;">
                            ${col.name}
                        </button>
                    `).join('')}
                </div>
            </div>
        </div>
    `;

    const temp = document.createElement('div');
    temp.innerHTML = html;
    document.body.appendChild(temp.firstElementChild);

    const modal = document.getElementById('addToCollectionModal');
    modal.querySelector('.modal-close').onclick = () => modal.remove();
    modal.onclick = (e) => {
        if (e.target === modal) modal.remove();
    };
}

// Add photo to collection
async function addPhotoToCollection(collectionId) {
    try {
        const response = await fetch(appPath(`/api/collections/${collectionId}/add`), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ photo_id: app.selectedPhoto.id })
        });

        if (response.ok) {
            showAlert('Photo added to collection', 'success');
            document.getElementById('addToCollectionModal')?.remove();
        }
    } catch (error) {
        console.error('Error adding to collection:', error);
        showAlert('Error adding to collection', 'error');
    }
}

async function toggleLike(photoId) {
    const photo = app.photos.find(item => item.id === photoId) || app.selectedPhoto;
    const liked = Boolean(photo && photo.liked_by_user);
    try {
        const response = await fetch(appPath(`/api/images/${photoId}/like`), {
            method: liked ? 'DELETE' : 'POST'
        });

        if (response.ok) {
            const state = await response.json();
            updatePhotoLikeState(photoId, state.like_count, state.liked_by_user);
            showAlert(state.liked_by_user ? 'Liked' : 'Like removed', 'success');
            return;
        }

        if (response.status === 401) {
            showAlert('Sign in to like media', 'info');
            return;
        }

        throw new Error(await response.text());
    } catch (error) {
        console.error('Error updating like:', error);
        showAlert('Error updating like', 'error');
    }
}

function updatePhotoLikeState(photoId, likeCount, likedByUser) {
    app.photos = app.photos.map(photo => {
        if (photo.id !== photoId) return photo;
        return { ...photo, like_count: likeCount, liked_by_user: likedByUser };
    });
    if (app.selectedPhoto && app.selectedPhoto.id === photoId) {
        app.selectedPhoto.like_count = likeCount;
        app.selectedPhoto.liked_by_user = likedByUser;
        updateModalLikeButton(app.selectedPhoto);
    }
    renderPhotos();
}

function updateModalLikeButton(photo) {
    const button = document.getElementById('likePhotoBtn');
    const text = document.getElementById('modalLikeText');
    if (!button || !text || !photo) return;
    button.classList.toggle('liked', Boolean(photo.liked_by_user));
    button.querySelector('.like-icon').textContent = photo.liked_by_user ? '♥' : '♡';
    text.textContent = `${photo.liked_by_user ? 'Unlike' : 'Like'} (${photo.like_count || 0})`;
}

// Utility functions
function debounce(func, delay) {
    let timeoutId;
    return function(...args) {
        clearTimeout(timeoutId);
        timeoutId = setTimeout(() => func(...args), delay);
    };
}

function showAlert(message, type = 'info') {
    const alert = document.createElement('div');
    alert.className = `alert alert-${type}`;
    alert.textContent = message;
    alert.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 2000;
        min-width: 300px;
        animation: slideIn 0.3s ease-out;
    `;

    document.body.appendChild(alert);

    setTimeout(() => {
        alert.style.animation = 'slideOut 0.3s ease-in';
        setTimeout(() => alert.remove(), 300);
    }, 3000);
}

function capitalizeFirstWord(value) {
    const title = String(value || '');
    return title.replace(/^(\s*)(\S)/, (_, leadingSpace, firstLetter) => leadingSpace + firstLetter.toUpperCase());
}

function escapeHtml(value) {
    return String(value ?? '').replace(/[&<>"']/g, (char) => ({
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
    }[char]));
}

// Animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(400px);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);
