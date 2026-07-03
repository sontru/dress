# Photo Library Web Application

A modern photo library web application built with Golang and SQLite, featuring photo discovery, categorization, collection management, and shopping cart functionality. The UI design is inspired by My Media Hub, offering a clean and intuitive interface for browsing and purchasing photos.

## Features

### Core Features
- **Photo Gallery**: Browse and search photos with pagination
- **Advanced Search**: Search by title, description, tags, and keywords
- **Categorization**: Organize photos by categories
- **Tagging System**: Filter photos by multiple tags
- **Collections**: Create and manage personal photo collections
- **Shopping Cart**: Add photos to cart and prepare for checkout
- **Responsive Design**: Works seamlessly on desktop, tablet, and mobile devices

### Technical Features
- RESTful API backend
- SQLite database with optimized queries
- Modular Go architecture
- Clean separation of concerns (handlers, models, database)
- Static file serving
- JSON API responses

## Project Structure

```
photo-library/
├── main.go                 # Application entry point
├── go.mod                  # Go module definition
├── config/
│   └── config.go          # Configuration management
├── models/
│   └── models.go          # Data models
├── database/
│   └── db.go              # Database initialization and schema
├── handlers/
│   └── handlers.go        # HTTP request handlers
├── templates/
│   ├── index.html         # Main gallery page
│   └── photo-detail.html  # Photo detail page
└── static/
    ├── css/
    │   ├── style.css      # Main styles
    │   └── layout.css     # Layout utilities
    └── js/
        └── app.js         # Frontend JavaScript
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- A web browser

### Installation

1. **Clone or download the project**
```bash
cd photo-library
```

2. **Download dependencies**
```bash
go mod download
```

3. **Build the application**
```bash
go build
```

4. **Run the application**
```bash
./photo-library
```

The application will start on `http://localhost:8082`

### Configuration

You can configure the application using environment variables:

```bash
# Set custom port
export PORT=3000

# Set database file path
export DATABASE_URL=my_photos.db

# Set environment
export ENVIRONMENT=production
```

## API Endpoints

### Photos
- `GET /api/photos` - Get all photos with pagination
- `POST /api/photos` - Create a new photo
- `PUT /api/photos/{id}` - Update a photo
- `DELETE /api/photos/{id}` - Delete a photo

### Search & Filter
- `GET /api/search?q=query&category=cat&tags=tag1,tag2&page=1` - Search photos
- `GET /api/tags` - Get all tags
- `GET /api/categories` - Get all categories

### Collections
- `GET /api/collections?user_id=1` - Get user's collections
- `POST /api/collections` - Create a collection
- `GET /api/collections/{id}` - Get collection details
- `POST /api/collections/{id}/add` - Add photo to collection

### Cart
- `GET /api/cart?user_id=1` - Get cart items
- `POST /api/cart/add` - Add item to cart
- `DELETE /api/cart/remove/{id}` - Remove item from cart

## Database Schema

### Tables

**photos** - Stores photo information
```sql
id, title, description, image_path, thumbnail, category, 
dimensions, file_type, file_size, orientation, resolution, 
color_mode, photographer, member_since, price, created_at, updated_at
```

**tags** - Photo tags/keywords
```sql
id, name
```

**photo_tags** - Junction table for photos and tags
```sql
photo_id, tag_id
```

**categories** - Photo categories
```sql
id, name, icon
```

**collections** - User collections
```sql
id, name, description, user_id, created_at, updated_at
```

**collection_photos** - Photos in collections
```sql
collection_id, photo_id, added_at
```

**cart_items** - Shopping cart
```sql
id, user_id, photo_id, quantity, price, added_at
```

**users** - User accounts
```sql
id, email, name, created_at
```

## Frontend Features

### Photo Grid
- Responsive grid layout (auto-fill with minimum column width)
- Hover effects and smooth transitions
- Click to view photo details

### Search and Filter
- Real-time search input
- Category filtering
- Multi-tag filtering
- Price range slider
- Pagination controls

### Photo Detail Modal
- Full photo view
- Detailed metadata display
- Photographer information
- Keywords/tags display
- Add to collection button
- Add to cart button

### Collections
- Create new collections
- View collection contents
- Add photos to collections
- Collections sidebar with quick access

### Shopping Cart
- View cart items
- Remove items
- Display total price
- Quantity tracking

## Styling

The application uses a modern, clean design with:
- CSS Grid for layout
- CSS Flexbox for components
- Custom CSS properties for theming
- Responsive breakpoints for mobile, tablet, and desktop
- Smooth animations and transitions

### Color Scheme
- Primary: `#007bff` (Blue)
- Secondary: `#6c757d` (Gray)
- Success: `#28a745` (Green)
- Danger: `#dc3545` (Red)
- Light Background: `#f8f9fa`

## Development

### Adding New Features

1. **Add a new handler**: Create functions in `handlers/handlers.go`
2. **Add API routes**: Register routes in `main.go`
3. **Add database models**: Define structs in `models/models.go`
4. **Add database schema**: Update `database/db.go`
5. **Update frontend**: Modify templates and JavaScript

### Sample Data

To insert sample photos into the database, you can use the provided `InsertSampleData` function:

```go
database.InsertSampleData(db)
```

Or manually insert photos via the API:

```bash
curl -X POST http://localhost:8082/api/photos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Woman in Elegant Dress",
    "description": "A stunning satin evening dress in champagne gold tone...",
    "image_path": "/images/dress-1.jpg",
    "thumbnail": "/images/dress-1-thumb.jpg",
    "category": "dress",
    "dimensions": "3000 x 4000 px",
    "file_type": "JPG",
    "file_size": "4.8 MB",
    "orientation": "Portrait",
    "resolution": "300 DPI",
    "color_mode": "RGB",
    "photographer": "Alex Morgan",
    "member_since": "May 2016",
    "price": 29.99
  }'
```

## Performance Considerations

- Database indexes on frequently queried columns (category, tags)
- Pagination to limit results per request
- SQLite for simple, single-file deployment
- Static file caching headers
- Debounced search input

## Future Enhancements

- [ ] User authentication and authorization
- [ ] Advanced image processing (cropping, effects)
- [ ] Image upload functionality
- [ ] Payment gateway integration
- [ ] Order management system
- [ ] User reviews and ratings
- [ ] Advanced analytics
- [ ] Batch operations
- [ ] Image optimization and CDN support
- [ ] WebP/AVIF format support

## Troubleshooting

### Application won't start
- Check if port 8082 is already in use
- Ensure Go is installed correctly
- Check database file path

### Database errors
- Delete `photo_library.db` and restart to reset database
- Check file permissions in the application directory

### Images not loading
- Ensure image paths in database match actual file locations
- Check `static` directory exists with image files

## License

This project is provided as a scaffold and can be freely modified and extended.

## Support

For issues or questions, refer to the comments in the source code or review the API documentation above.
