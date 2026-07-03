# Quick Start Guide

Get the Photo Library application up and running in minutes!

## Option 1: Using Make (Recommended)

### Requirements
- Go 1.21 or higher

### Steps

```bash
# 1. Navigate to the project directory
cd photo-library

# 2. Install dependencies and build
make install

# 3. Run the application
make run
```

The application will start on `http://localhost:8082`

## Option 2: Manual Build

```bash
# 1. Download dependencies
go mod download

# 2. Build the application
go build -o photo-library

# 3. Run the executable
./photo-library
```

## Option 3: Using Docker

### Requirements
- Docker
- Docker Compose

### Steps

```bash
# 1. Build and start the container
docker-compose up --build

# 2. Access the application
# Open http://localhost:8082 in your browser
```

To stop: Press `Ctrl+C`

## Option 4: Development Mode (Hot Reload)

### Requirements
- Go 1.21 or higher
- Air (for live reload)

```bash
# 1. Install Air
go install github.com/cosmtrek/air@latest

# 2. Run with hot reload
make dev
```

Changes to Go files will automatically rebuild and restart the server.

## Accessing the Application

1. Open your browser and navigate to: **http://localhost:8082**

2. You should see the Photo Library homepage with:
   - Navigation menu at the top
   - Search bar
   - Photo grid (with sample data)
   - Filters sidebar (categories, tags, price)
   - Collections sidebar

## First Steps

### Browse Photos
- Click on any photo card to view details
- See full metadata, photographer info, and tags

### Search Photos
- Use the search bar at the top to find photos
- Filter by category using the left sidebar
- Select tags to filter results
- Use the price slider to limit price range

### Create Collections
- Click "+ New" in the Collections sidebar
- Enter a name and description
- Click "Create" to add the collection
- Click on a photo and "Add to collection" to add photos

### Add to Cart
- Click on a photo to view details
- Click "Add to cart" button
- View cart using the cart icon in the header

## Configuration

Edit environment variables before running:

```bash
# Custom port (default: 8082)
export PORT=3000

# Custom database location
export DATABASE_URL=my_database.db

# Environment type
export ENVIRONMENT=development
```

## Stopping the Application

- Press `Ctrl+C` in the terminal where the app is running
- Or use `docker-compose down` if running in Docker

## Troubleshooting

### Port already in use
```bash
# Use a different port
export PORT=3001
./photo-library
```

### Database errors
```bash
# Reset database (deletes all data!)
rm photo_library.db
./photo-library
```

### Missing images
- Sample images are referenced but not included
- You can use placeholder services or add your own images to `static/images/`

## Next Steps

- See [README.md](README.md) for detailed API documentation
- Check [main.go](main.go) for code structure
- Review [handlers/handlers.go](handlers/handlers.go) for API implementations
- Explore [static/js/app.js](static/js/app.js) for frontend logic

## Support

For detailed information, refer to the main README.md file in the project root.
