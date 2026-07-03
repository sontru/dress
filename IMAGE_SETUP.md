# Image Setup & Management Guide

## Current Setup: Placeholder Images

The application is currently using **placeholder images** from `picsum.photos` service:
- **Advantages**: Works immediately, no setup needed, high-quality random images
- **Disadvantages**: Images change randomly, not customizable, requires internet

## How to Use Real Images

### Option 1: Upload Images to Local Storage (Recommended for Development)

#### Step 1: Prepare Your Images

```bash
# Create a directory for your images
mkdir -p /home/sont/Store/dress/static/images

# Add your images to this directory
# Example:
cp /path/to/your/photo1.jpg /home/sont/Store/dress/static/images/
cp /path/to/your/photo2.jpg /home/sont/Store/dress/static/images/
```

#### Step 2: Update seed.go

Replace placeholder URLs with local paths:

**Before (Placeholder):**
```go
imagePath:    "https://picsum.photos/3000/4000?random=1",
thumbnail:    "https://picsum.photos/300/400?random=1",
```

**After (Local):**
```go
imagePath:    "/static/images/photo1.jpg",
thumbnail:    "/static/images/photo1-thumb.jpg",
```

#### Step 3: Create Thumbnails

Generate thumbnail versions of your images:

```bash
# Using ImageMagick
convert /home/sont/Store/dress/static/images/photo1.jpg \
        -resize 300x400 \
        /home/sont/Store/dress/static/images/photo1-thumb.jpg

# Or install ImageMagick first:
sudo apt-get install imagemagick
```

#### Step 4: Rebuild and Restart

```bash
cd /home/sont/Store/dress
rm -f photo_library.db
go build -o photo-library
./photo-library
```

---

### Option 2: Use External CDN (Production)

#### Step 1: Upload Images to CDN

Popular services:
- **AWS S3** - `https://your-bucket.s3.amazonaws.com/photo1.jpg`
- **Cloudinary** - `https://res.cloudinary.com/your-cloud/image/upload/photo1.jpg`
- **Firebase Storage** - `https://firebasestorage.googleapis.com/v0/b/bucket/o/photo1.jpg`
- **Imgur** - `https://imgur.com/abcd123.jpg`

#### Step 2: Update seed.go

```go
imagePath:    "https://your-cdn.com/images/photo1.jpg",
thumbnail:    "https://your-cdn.com/images/photo1-thumb.jpg",
```

#### Step 3: Rebuild

```bash
cd /home/sont/Store/dress
rm -f photo_library.db
go build -o photo-library
./photo-library
```

---

### Option 3: API Upload (Production Ready)

Add image upload endpoint:

#### Step 1: Create Upload Handler

```go
// Add to handlers/handlers.go
import "io"

func UploadPhotoHandler(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := r.ParseMultipartForm(10 << 20) // 10MB limit
        if err != nil {
            http.Error(w, "File too large", http.StatusBadRequest)
            return
        }

        file, handler, err := r.FormFile("image")
        if err != nil {
            http.Error(w, "Error uploading file", http.StatusBadRequest)
            return
        }
        defer file.Close()

        // Save file
        filename := filepath.Join("static/images", handler.Filename)
        dst, err := os.Create(filename)
        if err != nil {
            http.Error(w, "Error saving file", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        io.Copy(dst, file)

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "path": "/static/images/" + handler.Filename,
        })
    }
}
```

#### Step 2: Add Route in main.go

```go
apiRouter.HandleFunc("/upload", handlers.UploadPhotoHandler(db)).Methods("POST")
```

#### Step 3: Use API to Upload

```bash
curl -X POST http://localhost:8082/api/upload \
  -F "image=@/path/to/photo.jpg"
```

---

## Image Size Recommendations

### Optimal Dimensions

| Type | Width | Height | DPI | Format |
|------|-------|--------|-----|--------|
| Full Photo | 3000 | 4000 | 300 | JPG |
| Thumbnail | 300 | 400 | 72 | JPG |
| Mobile | 600 | 800 | 72 | JPG |

### File Size Guidelines

| Quality | File Size | Use Case |
|---------|-----------|----------|
| High (Premium) | 4-5 MB | Full resolution downloads |
| Medium | 1-2 MB | Web display |
| Low (Thumbnail) | 50-100 KB | Thumbnails & previews |

---

## Batch Image Management

### Add Multiple Images Script

Create `batch-import.sh`:

```bash
#!/bin/bash

SOURCE_DIR="${1:-.}"
DEST_DIR="/home/sont/Store/dress/static/images"

mkdir -p "$DEST_DIR"

for img in "$SOURCE_DIR"/*.{jpg,jpeg,png,webp}; do
    if [ -f "$img" ]; then
        echo "Processing: $(basename "$img")"
        
        # Copy original
        cp "$img" "$DEST_DIR/"
        
        # Create thumbnail
        convert "$img" -resize 300x400 "$DEST_DIR/$(basename "$img" )-thumb.jpg"
        
        echo "✓ Done"
    fi
done

echo "All images processed!"
```

Usage:
```bash
chmod +x batch-import.sh
./batch-import.sh /path/to/your/images/folder
```

---

## Database Operations

### Add Photo Manually via API

```bash
curl -X POST http://localhost:8082/api/photos \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My Photo",
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
    "photographer": "Your Name",
    "member_since": "January 2024",
    "price": 29.99
  }'
```

### Query Photos from Database

```bash
# Open SQLite
sqlite3 /home/sont/Store/dress/photo_library.db

# View all photos
SELECT id, title, image_path, price FROM photos;

# Update image path
UPDATE photos SET image_path = '/static/images/newphoto.jpg' WHERE id = 1;

# Delete photo
DELETE FROM photos WHERE id = 1;
```

---

## Switching Between Placeholder and Real Images

### Keep Placeholder for Development

```go
// In seed.go, use conditional compilation
//go:build debug
// +build debug

// ... placeholder image setup
```

Build with debug:
```bash
go build -tags debug -o photo-library
```

### Use Real Images for Production

```go
// In seed.go, default to real images
// ... real image setup
```

Build normally:
```bash
go build -o photo-library
```

---

## Performance Tips

### Image Optimization

```bash
# Compress JPEG
ffmpeg -i input.jpg -q:v 5 output.jpg

# Convert to WebP (better compression)
ffmpeg -i input.jpg -c:v libwebp -q:v 80 output.webp

# Resize image
ffmpeg -i input.jpg -vf scale=3000:4000 output.jpg
```

### Caching

Add HTTP caching headers in main.go:

```go
router.PathPrefix("/static/").Handler(
    http.FileServer(http.Dir("static")),
)

// Add middleware for caching
func cacheMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Cache-Control", "public, max-age=86400")
        next.ServeHTTP(w, r)
    })
}
```

---

## Troubleshooting

### Images Not Loading

```bash
# Check if files exist
ls -la /home/sont/Store/dress/static/images/

# Check file permissions
chmod 644 /home/sont/Store/dress/static/images/*.jpg

# Check database paths
sqlite3 /home/sont/Store/dress/photo_library.db "SELECT image_path FROM photos LIMIT 1;"
```

### Slow Image Loading

- Reduce file size (compress with ffmpeg)
- Use CDN for production
- Enable caching headers
- Use WebP format instead of JPG

### CORS Issues (if using external CDN)

Add CORS middleware:

```go
router.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        next.ServeHTTP(w, r)
    })
})
```

---

## Quick Start Commands

```bash
# Setup local images
mkdir -p /home/sont/Store/dress/static/images
cp your-photos.jpg /home/sont/Store/dress/static/images/

# Generate thumbnails
convert your-photos.jpg -resize 300x400 your-photos-thumb.jpg

# Reset database and rebuild
cd /home/sont/Store/dress
rm -f photo_library.db
go build -o photo-library

# Run application
./photo-library

# Visit in browser
# http://localhost:8082
```

---

## Support

For detailed image handling, refer to:
- [ImageMagick Documentation](https://imagemagick.org/)
- [FFmpeg Documentation](https://ffmpeg.org/documentation.html)
- [WebP Format](https://developers.google.com/speed/webp)
