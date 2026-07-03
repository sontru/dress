// App state
const app = {
    currentPage: 1,
    pageSize: 'all',
    photos: [],
    totalPhotos: 0,
    selectedPhoto: null,
    cart: [],
    collections: [],
    filters: {
        category: null,
        tags: [],
        priceMax: 100,
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
const cartCountEl = document.getElementById('cartCount');

// Initialize app
document.addEventListener('DOMContentLoaded', () => {
    loadPhotos();
    loadCategories();
    loadTags();
    loadCollections();
    setupEventListeners();
    loadCart();
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

    // Price range
    document.getElementById('priceRange').addEventListener('input', (e) => {
        app.filters.priceMax = e.target.value;
        document.getElementById('priceDisplay').textContent = `$0 - $${e.target.value}`;
        app.currentPage = 1;
        loadPhotos();
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

    // Cart link
    document.querySelector('.cart-link').addEventListener('click', (e) => {
        e.preventDefault();
        showCart();
    });
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
                window.location.href = appPath('/');
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

        // If search or filters applied, use search endpoint
        if (app.filters.searchQuery || app.filters.tags.length > 0) {
            url = `/api/search?q=${encodeURIComponent(app.filters.searchQuery)}`;
            if (app.filters.category) url += `&category=${app.filters.category}`;
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
        console.error('Error loading photos:', error);
        showAlert('Error loading photos', 'error');
    }
}

// Render photos grid
function renderPhotos() {
    if (app.photos.length === 0) {
        photosGrid.innerHTML = `
            <div class="empty-state" style="grid-column: 1 / -1;">
                <div class="empty-state-icon">📷</div>
                <h3>No photos found</h3>
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
                <div class="photo-card-price">$${photo.price || '29.99'}</div>
            </div>
        </div>
    `).join('');
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
    window.location.href = appPath(`/photo/${photoId}`);
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

// Cart functions
async function loadCart() {
    try {
        const response = await fetch(appPath('/api/cart?user_id=1'));
        const data = await response.json();
        app.cart = data.items || [];
        updateCartCount();
    } catch (error) {
        console.error('Error loading cart:', error);
    }
}

function updateCartCount() {
    cartCountEl.textContent = app.cart.length;
}

async function addToCart(photoId) {
    try {
        const response = await fetch(appPath('/api/cart/add'), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                user_id: 1,
                photo_id: photoId
            })
        });

        if (response.ok) {
            loadCart();
            showAlert('Photo added to cart', 'success');
        }
    } catch (error) {
        console.error('Error adding to cart:', error);
        showAlert('Error adding to cart', 'error');
    }
}

function showCart() {
    const total = app.cart.reduce((sum, item) => sum + (item.price * item.quantity), 0);
    
    const html = `
        <div class="modal active" id="cartModal">
            <div class="modal-content">
                <button class="modal-close">&times;</button>
                <h2 style="padding: 20px;">Shopping Cart</h2>
                <div style="padding: 0 20px;">
                    ${app.cart.length === 0 ? 
                        '<p class="empty-state">Your cart is empty</p>' :
                        `<div>
                            ${app.cart.map(item => `
                                <div style="padding: 15px 0; border-bottom: 1px solid #e9ecef; display: flex; justify-content: space-between; align-items: center;">
                                    <div>
                                        <p style="margin: 0; font-weight: 500;">Photo #${item.photo_id}</p>
                                        <p style="margin: 5px 0 0; color: #6c757d; font-size: 13px;">Qty: ${item.quantity}</p>
                                    </div>
                                    <div style="text-align: right;">
                                        <p style="margin: 0; font-weight: 600;">$${(item.price * item.quantity).toFixed(2)}</p>
                                        <button onclick="removeFromCart(${item.id})" style="margin-top: 5px; padding: 4px 8px; background: #dc3545; color: white; border: none; border-radius: 4px; cursor: pointer; font-size: 12px;">Remove</button>
                                    </div>
                                </div>
                            `).join('')}
                        </div>`
                    }
                    ${app.cart.length > 0 ?
                        `<div style="padding: 20px 0; border-top: 2px solid #e9ecef; display: flex; justify-content: space-between; align-items: center; margin-top: 20px;">
                            <h3 style="margin: 0;">Total:</h3>
                            <p style="margin: 0; font-size: 24px; font-weight: 700; color: #007bff;">$${total.toFixed(2)}</p>
                        </div>
                        <button class="btn-primary" style="width: 100%; margin-top: 20px; padding: 12px;">Proceed to Checkout</button>` :
                        ''
                    }
                </div>
            </div>
        </div>
    `;

    const temp = document.createElement('div');
    temp.innerHTML = html;
    document.body.appendChild(temp.firstElementChild);

    const modal = document.getElementById('cartModal');
    modal.querySelector('.modal-close').onclick = () => modal.remove();
    modal.onclick = (e) => {
        if (e.target === modal) modal.remove();
    };
}

async function removeFromCart(itemId) {
    try {
        const response = await fetch(appPath(`/api/cart/remove/${itemId}`), {
            method: 'DELETE'
        });

        if (response.ok) {
            loadCart();
            // Refresh cart modal if open
            const cartModal = document.getElementById('cartModal');
            if (cartModal) {
                showCart();
            }
            showAlert('Item removed from cart', 'success');
        }
    } catch (error) {
        console.error('Error removing from cart:', error);
        showAlert('Error removing from cart', 'error');
    }
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
