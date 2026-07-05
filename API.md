# API Documentation

Complete API reference for the Photo Library application.

## Base URL

```
http://localhost:8082/api
```

## Response Format

All API responses are in JSON format with the following structure:

### Success Response
```json
{
  "data": { ... },
  "status": "success"
}
```

### Error Response
```json
{
  "error": "Error message",
  "status": "error",
  "code": 400
}
```

## Endpoints

### Photos

#### Get All Photos
```http
GET /api/photos?page=1&pageSize=12
```

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `pageSize` (integer): Items per page (default: 12)

**Response:**
```json
{
  "photos": [
    {
      "id": 1,
      "title": "Woman in Elegant Dress",
      "description": "...",
      "image_path": "/static/images/dress-1.jpg",
      "thumbnail": "/static/images/dress-1-thumb.jpg",
      "category": "dress",
      "dimensions": "3000 x 4000 px",
      "file_type": "JPG",
      "file_size": "4.8 MB",
      "orientation": "Portrait",
      "resolution": "300 DPI",
      "color_mode": "RGB",
      "photographer": "Alex Morgan",
      "member_since": "May 2016",
      "like_count": 0,
      "liked_by_user": false,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "page_size": 12
}
```

#### Create Photo
```http
POST /api/photos
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Photo Title",
  "description": "Photo description",
  "image_path": "/static/images/photo.jpg",
  "thumbnail": "/static/images/photo-thumb.jpg",
  "category": "dress",
  "dimensions": "3000 x 4000 px",
  "file_type": "JPG",
  "file_size": "4.8 MB",
  "orientation": "Portrait",
  "resolution": "300 DPI",
  "color_mode": "RGB",
  "photographer": "Photographer Name"
}
```

`member_since` values in photo responses are derived from the owning user's `created_at` timestamp.

**Response:**
```json
{
  "id": 123
}
```

#### Update Photo
```http
PUT /api/photos/{id}
Content-Type: application/json
```

**Request Body:**
```json
{
  "title": "Updated Title",
  "description": "Updated description",
  "category": "dress"
}
```

**Response:**
```json
{
  "message": "Photo updated successfully"
}
```

#### Delete Photo
```http
DELETE /api/photos/{id}
```

**Response:**
```json
{
  "message": "Photo deleted successfully"
}
```

### Search & Filter

#### Search Photos
```http
GET /api/search?q=query&category=dress&tags=elegant,satin&page=1
```

**Query Parameters:**
- `q` (string): Search query
- `category` (string): Filter by category
- `tags` (string): Comma-separated tags
- `page` (integer): Page number

**Response:** Same as Get All Photos

#### Get Tags
```http
GET /api/tags
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "elegant"
  },
  {
    "id": 2,
    "name": "satin"
  }
]
```

#### Get Categories
```http
GET /api/categories
```

**Response:**
```json
[
  {
    "id": 1,
    "name": "dress",
    "icon": "👗"
  },
  {
    "id": 2,
    "name": "fashion",
    "icon": "👔"
  }
]
```

### Collections

#### Get User Collections
```http
GET /api/collections?user_id=1
```

**Query Parameters:**
- `user_id` (integer): User ID

**Response:**
```json
[
  {
    "id": 1,
    "name": "Favorites",
    "description": "My favorite photos",
    "user_id": 1,
    "photo_count": 5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
]
```

#### Create Collection
```http
POST /api/collections
Content-Type: application/json
```

**Request Body:**
```json
{
  "name": "Summer Collection",
  "description": "Photos from summer 2024",
  "user_id": 1
}
```

**Response:**
```json
{
  "id": 2
}
```

#### Get Collection Details
```http
GET /api/collections/{id}
```

**Response:**
```json
{
  "collection": {
    "id": 1,
    "name": "Favorites",
    "description": "My favorite photos",
    "user_id": 1,
    "photo_count": 5,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  },
  "photos": [
    {
      "id": 1,
      "title": "Photo Title",
      "image_path": "/static/images/photo.jpg",
      "thumbnail": "/static/images/photo-thumb.jpg",
      "like_count": 0,
      "liked_by_user": false
    }
  ]
}
```

#### Add Photo to Collection
```http
POST /api/collections/{id}/add
Content-Type: application/json
```

**Request Body:**
```json
{
  "photo_id": 1
}
```

**Response:**
```json
{
  "message": "Photo added to collection"
}
```

### Likes

#### Like Image
```http
POST /api/images/{id}/like
```

**Response:**
```json
{
  "image_id": 5,
  "like_count": 12,
  "liked_by_user": true
}
```

#### Unlike Image
```http
DELETE /api/images/{id}/like
```

**Response:**
```json
{
  "image_id": 5,
  "like_count": 11,
  "liked_by_user": false
}
```

## Error Codes

| Code | Meaning |
|------|---------|
| 200 | OK - Request succeeded |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid parameters |
| 404 | Not Found - Resource not found |
| 500 | Internal Server Error - Server error |

## Rate Limiting

Currently, the API has no rate limiting. In production, implement rate limiting based on:
- IP address
- User authentication
- API key

## Authentication

Currently, the API has no authentication. Implement JWT or API keys for production.

## CORS

Currently, CORS is not enabled. Add CORS middleware for cross-origin requests.

## Pagination

All list endpoints support pagination:
- Default page size: 12
- Maximum page size: 100 (recommended)

## Filtering

### By Category
```http
GET /api/search?category=dress
```

### By Tags
```http
GET /api/search?tags=elegant,satin
```

### By Search Query
```http
GET /api/search?q=dress
```

## Sorting

Currently, results are ordered by creation date (newest first).

To add sorting:
```http
GET /api/photos?sort=likes&order=desc
```

## Example Requests

### Search for elegant photos
```bash
curl "http://localhost:8082/api/search?q=elegant&category=dress"
```

### Get user's collections
```bash
curl "http://localhost:8082/api/collections?user_id=1"
```

### Like an image
```bash
curl -X POST "http://localhost:8082/api/images/5/like"
```

### Create a collection
```bash
curl -X POST "http://localhost:8082/api/collections" \
  -H "Content-Type: application/json" \
  -d '{"name": "Summer", "description": "Summer photos", "user_id": 1}'
```

## Testing the API

Use curl, Postman, or any HTTP client:

```bash
# Get all photos
curl http://localhost:8082/api/photos

# Search photos
curl "http://localhost:8082/api/search?q=elegant"

# Get tags
curl http://localhost:8082/api/tags

# Get categories
curl http://localhost:8082/api/categories
```
