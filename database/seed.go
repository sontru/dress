package database

import (
	"database/sql"
	"log"
)

// SeedDatabase inserts sample data into the database
func SeedDatabase(db *sql.DB) error {
	// Insert sample user
	_, err := db.Exec(
		"INSERT OR IGNORE INTO users (id, email, name) VALUES (?, ?, ?)",
		1, "user@example.com", "Sample User")
	if err != nil {
		return err
	}

	// Insert sample categories
	categories := []struct {
		name string
		icon string
	}{
		{"Photos", "📷"},
		{"Illustrations", "🎨"},
		{"Vectors", "📐"},
		{"Videos", "🎬"},
		{"dress", "👗"},
		{"evening dress", "🌙"},
		{"fashion", "👔"},
		{"portraits", "👤"},
	}

	for _, cat := range categories {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO categories (name, icon) VALUES (?, ?)",
			cat.name, cat.icon)
		if err != nil {
			log.Printf("Error inserting category: %v", err)
		}
	}

	// Insert sample tags
	tags := []string{
		"elegant", "satin", "champagne", "luxury", "formal",
		"backless", "slit", "party", "wedding", "glamour", "portrait",
		"dress", "fashion", "woman", "style", "evening",
	}

	for _, tag := range tags {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO tags (name) VALUES (?)",
			tag)
		if err != nil {
			log.Printf("Error inserting tag: %v", err)
		}
	}

	// Insert sample photos
	photos := []struct {
		title        string
		description  string
		imagePath    string
		thumbnail    string
		category     string
		dimensions   string
		fileType     string
		fileSize     string
		orientation  string
		resolution   string
		colorMode    string
		photographer string
		memberSince  string
		price        float64
	}{
		{
			title:        "Woman in Elegant Dress",
			description:  "A stunning satin evening dress in a champagne gold tone with a sleek, fitted silhouette. Crafted from luxurious satin fabric that drapes beautifully and feels smooth on the skin. Features a low back, delicate spaghetti straps, and a high side slit for an elegant look. Perfect for formal events, weddings, galas, proms, evening parties and red carpet occasions.",
			imagePath:    "https://picsum.photos/3000/4000?random=1",
			thumbnail:    "https://picsum.photos/300/400?random=1",
			category:     "dress",
			dimensions:   "3000 x 4000 px",
			fileType:     "JPG",
			fileSize:     "4.8 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Alex Morgan",
			memberSince:  "May 2016",
			price:        29.99,
		},
		{
			title:        "Evening Gown",
			description:  "Beautiful evening gown with flowing fabric and elegant design.",
			imagePath:    "https://picsum.photos/2560/3840?random=2",
			thumbnail:    "https://picsum.photos/300/400?random=2",
			category:     "evening dress",
			dimensions:   "2560 x 3840 px",
			fileType:     "JPG",
			fileSize:     "3.2 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Sarah Johnson",
			memberSince:  "March 2018",
			price:        34.99,
		},
		{
			title:        "Fashion Portrait",
			description:  "Professional fashion photography showcasing the latest collection.",
			imagePath:    "https://picsum.photos/3000/4000?random=3",
			thumbnail:    "https://picsum.photos/300/400?random=3",
			category:     "fashion",
			dimensions:   "3000 x 4000 px",
			fileType:     "JPG",
			fileSize:     "5.1 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Michael Chen",
			memberSince:  "June 2019",
			price:        24.99,
		},
		{
			title:        "Studio Portrait",
			description:  "Professional studio portrait with perfect lighting and composition.",
			imagePath:    "https://picsum.photos/2000/3000?random=4",
			thumbnail:    "https://picsum.photos/300/400?random=4",
			category:     "portraits",
			dimensions:   "2000 x 3000 px",
			fileType:     "JPG",
			fileSize:     "2.8 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Emma Wilson",
			memberSince:  "January 2017",
			price:        19.99,
		},
		{
			title:        "Elegant Cocktail Dress",
			description:  "Sophisticated black cocktail dress perfect for evening events.",
			imagePath:    "https://picsum.photos/3000/4000?random=5",
			thumbnail:    "https://picsum.photos/300/400?random=5",
			category:     "dress",
			dimensions:   "3000 x 4000 px",
			fileType:     "JPG",
			fileSize:     "4.5 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "David Martinez",
			memberSince:  "February 2015",
			price:        24.99,
		},
		{
			title:        "Luxury Evening Wear",
			description:  "Premium evening wear collection with exquisite details.",
			imagePath:    "https://picsum.photos/2560/3840?random=6",
			thumbnail:    "https://picsum.photos/300/400?random=6",
			category:     "evening dress",
			dimensions:   "2560 x 3840 px",
			fileType:     "JPG",
			fileSize:     "3.8 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Isabella Rodriguez",
			memberSince:  "August 2014",
			price:        39.99,
		},
		{
			title:        "Modern Fashion Look",
			description:  "Contemporary fashion photography with urban backdrop.",
			imagePath:    "https://picsum.photos/3000/4000?random=7",
			thumbnail:    "https://picsum.photos/300/400?random=7",
			category:     "fashion",
			dimensions:   "3000 x 4000 px",
			fileType:     "JPG",
			fileSize:     "5.2 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "James Photography",
			memberSince:  "July 2016",
			price:        29.99,
		},
		{
			title:        "Professional Headshot",
			description:  "High-quality professional headshot for business use.",
			imagePath:    "https://picsum.photos/2000/3000?random=8",
			thumbnail:    "https://picsum.photos/300/400?random=8",
			category:     "portraits",
			dimensions:   "2000 x 3000 px",
			fileType:     "JPG",
			fileSize:     "2.6 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Lisa Chen",
			memberSince:  "November 2017",
			price:        16.99,
		},
		{
			title:        "Designer Evening Collection",
			description:  "Haute couture evening dress with intricate embellishments.",
			imagePath:    "https://picsum.photos/3000/4000?random=9",
			thumbnail:    "https://picsum.photos/300/400?random=9",
			category:     "evening dress",
			dimensions:   "3000 x 4000 px",
			fileType:     "JPG",
			fileSize:     "4.9 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Roberto Fontana",
			memberSince:  "December 2013",
			price:        44.99,
		},
		{
			title:        "Casual Elegance",
			description:  "Stylish casual wear with elegant finishing touches.",
			imagePath:    "https://picsum.photos/2560/3840?random=10",
			thumbnail:    "https://picsum.photos/300/400?random=10",
			category:     "fashion",
			dimensions:   "2560 x 3840 px",
			fileType:     "JPG",
			fileSize:     "3.4 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Nina Petrov",
			memberSince:  "April 2018",
			price:        21.99,
		},
		{
			title:        "Portrait in Natural Light",
			description:  "Beautiful portrait taken in natural daylight setting.",
			imagePath:    "https://picsum.photos/2000/3000?random=11",
			thumbnail:    "https://picsum.photos/300/400?random=11",
			category:     "portraits",
			dimensions:   "2000 x 3000 px",
			fileType:     "JPG",
			fileSize:     "2.7 MB",
			orientation:  "Portrait",
			resolution:   "300 DPI",
			colorMode:    "RGB",
			photographer: "Marcus Johnson",
			memberSince:  "September 2015",
			price:        18.99,
		},
	}

	for _, photo := range photos {
		result, err := db.Exec(
			`INSERT OR IGNORE INTO photos 
			 (title, description, image_path, thumbnail, category, dimensions,
			  file_type, file_size, orientation, resolution, color_mode, 
			  photographer, member_since, price)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			photo.title, photo.description, photo.imagePath, photo.thumbnail,
			photo.category, photo.dimensions, photo.fileType, photo.fileSize,
			photo.orientation, photo.resolution, photo.colorMode,
			photo.photographer, photo.memberSince, photo.price)

		if err != nil {
			log.Printf("Error inserting photo: %v", err)
			continue
		}

		photoID, err := result.LastInsertId()
		if err != nil {
			log.Printf("Error getting photo ID: %v", err)
			continue
		}

		// Add sample tags to each photo
		tagMapping := map[int][]int{
			1:  {1, 2, 3, 4, 5, 6, 7}, // Woman in Elegant Dress - elegant, satin, champagne, luxury, formal, backless, slit
			2:  {4, 5, 16, 9, 10},     // Evening Gown - luxury, formal, evening, wedding, glamour
			3:  {4, 13, 14, 15, 1},    // Fashion Portrait - luxury, fashion, woman, style, elegant
			4:  {14, 11, 10, 15},      // Studio Portrait - woman, portrait, glamour, style
			5:  {1, 5, 12, 2},         // Elegant Cocktail Dress - elegant, formal, dress, satin
			6:  {4, 16, 5, 1, 2},      // Luxury Evening Wear - luxury, evening, formal, elegant, satin
			7:  {13, 15, 1, 14},       // Modern Fashion Look - fashion, style, elegant, woman
			8:  {14, 11, 4, 15},       // Professional Headshot - woman, portrait, luxury, style
			9:  {4, 16, 5, 1, 2, 6},   // Designer Evening Collection - luxury, evening, formal, elegant, satin, backless
			10: {13, 15, 1, 5},        // Casual Elegance - fashion, style, elegant, formal
			11: {14, 11, 1, 15},       // Portrait in Natural Light - woman, portrait, elegant, style
		}

		if tags, ok := tagMapping[int(photoID)]; ok {
			for _, tagID := range tags {
				_, err := db.Exec(
					"INSERT OR IGNORE INTO photo_tags (photo_id, tag_id) VALUES (?, ?)",
					photoID, tagID)
				if err != nil {
					log.Printf("Error linking tag to photo: %v", err)
				}
			}
		}
	}

	// Insert sample collection
	result, err := db.Exec(
		"INSERT OR IGNORE INTO collections (name, description, user_id) VALUES (?, ?, ?)",
		"Favorites", "My favorite photos", 1)
	if err != nil {
		log.Printf("Error creating collection: %v", err)
	}

	collectionID, _ := result.LastInsertId()

	// Add photos to collection
	for i := 1; i <= 3; i++ {
		_, err := db.Exec(
			"INSERT OR IGNORE INTO collection_photos (collection_id, photo_id) VALUES (?, ?)",
			collectionID, i)
		if err != nil {
			log.Printf("Error adding photo to collection: %v", err)
		}
	}

	log.Println("Sample data seeded successfully!")
	return nil
}
