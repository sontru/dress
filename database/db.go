package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Init(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func CreateTables(db *sql.DB) error {
	if err := migrateLegacyImageRelationTables(db); err != nil {
		return err
	}

	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE,
		email TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		real_name TEXT,
		password_hash TEXT,
		bio TEXT,
		location TEXT,
		website TEXT,
		avatar_url TEXT,
		role TEXT DEFAULT 'member',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS images (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		image_path TEXT NOT NULL,
		thumbnail TEXT,
		category TEXT,
		user_category TEXT,
		dimensions TEXT,
		file_type TEXT,
		file_size TEXT,
		orientation TEXT,
		resolution TEXT,
		color_mode TEXT,
		photographer TEXT,
		captured_at TEXT,
		is_public INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS photos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		title TEXT NOT NULL,
		description TEXT,
		image_path TEXT NOT NULL,
		thumbnail TEXT,
		category TEXT DEFAULT 'Photos',
		user_category TEXT,
		file_name TEXT,
		dimensions TEXT,
		file_type TEXT,
		file_size TEXT,
		orientation TEXT,
		resolution TEXT,
		color_mode TEXT,
		photographer TEXT,
		captured_at TEXT,
		photo_location TEXT,
		camera TEXT,
		focal_length TEXT,
		is_public INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS image_tags (
		image_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		FOREIGN KEY (image_id) REFERENCES images(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id),
		PRIMARY KEY (image_id, tag_id)
	);

	CREATE TABLE IF NOT EXISTS photo_tags (
		photo_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		FOREIGN KEY (photo_id) REFERENCES photos(id),
		FOREIGN KEY (tag_id) REFERENCES tags(id),
		PRIMARY KEY (photo_id, tag_id)
	);

	CREATE TABLE IF NOT EXISTS keywords (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS image_keywords (
		image_id INTEGER NOT NULL,
		keyword_id INTEGER NOT NULL,
		FOREIGN KEY (image_id) REFERENCES images(id),
		FOREIGN KEY (keyword_id) REFERENCES keywords(id),
		PRIMARY KEY (image_id, keyword_id)
	);

	CREATE TABLE IF NOT EXISTS photo_keywords (
		photo_id INTEGER NOT NULL,
		keyword_id INTEGER NOT NULL,
		FOREIGN KEY (photo_id) REFERENCES photos(id),
		FOREIGN KEY (keyword_id) REFERENCES keywords(id),
		PRIMARY KEY (photo_id, keyword_id)
	);

	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL,
		icon TEXT
	);

	CREATE TABLE IF NOT EXISTS user_categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		UNIQUE (user_id, name)
	);

	CREATE TABLE IF NOT EXISTS collections (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		user_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS collection_images (
		collection_id INTEGER NOT NULL,
		image_id INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (collection_id) REFERENCES collections(id),
		FOREIGN KEY (image_id) REFERENCES images(id),
		PRIMARY KEY (collection_id, image_id)
	);

	CREATE TABLE IF NOT EXISTS collection_photos (
		collection_id INTEGER NOT NULL,
		photo_id INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (collection_id) REFERENCES collections(id),
		FOREIGN KEY (photo_id) REFERENCES photos(id),
		PRIMARY KEY (collection_id, photo_id)
	);

	CREATE TABLE IF NOT EXISTS image_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		image_id INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (image_id) REFERENCES images(id),
		UNIQUE (user_id, image_id)
	);

	CREATE TABLE IF NOT EXISTS photo_likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		photo_id INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (photo_id) REFERENCES photos(id),
		UNIQUE (user_id, photo_id)
	);

	CREATE INDEX IF NOT EXISTS idx_images_category ON images(category);
	CREATE INDEX IF NOT EXISTS idx_photos_category ON photos(category);
	CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id);
	CREATE INDEX IF NOT EXISTS idx_image_tags_tag_id ON image_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_photo_tags_tag_id ON photo_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id);
	CREATE INDEX IF NOT EXISTS idx_image_likes_image_id ON image_likes(image_id);
	CREATE INDEX IF NOT EXISTS idx_image_likes_user_id ON image_likes(user_id);
	CREATE INDEX IF NOT EXISTS idx_photo_likes_photo_id ON photo_likes(photo_id);
	CREATE INDEX IF NOT EXISTS idx_photo_likes_user_id ON photo_likes(user_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := migrateUsersTable(db); err != nil {
		return err
	}
	if err := migrateImagesTable(db); err != nil {
		return err
	}
	if err := migratePhotosTable(db); err != nil {
		return err
	}
	if err := copyPhotoMediaFromImages(db); err != nil {
		return err
	}
	if err := dropLegacyMemberSinceColumns(db); err != nil {
		return err
	}
	if err := dropLegacyImagePriceColumn(db); err != nil {
		return err
	}
	if err := dropLegacyImagePhotoMetadataColumns(db); err != nil {
		return err
	}
	if err := createPostMigrationIndexes(db); err != nil {
		return err
	}

	return nil
}

func createPostMigrationIndexes(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_images_user_id ON images(user_id)",
		"CREATE INDEX IF NOT EXISTS idx_user_categories_user_id ON user_categories(user_id)",
	}

	for _, statement := range indexes {
		if _, err := db.Exec(statement); err != nil {
			return fmt.Errorf("failed to create post-migration index: %w", err)
		}
	}

	return nil
}

func migrateUsersTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"username", "TEXT"},
		{"real_name", "TEXT"},
		{"password_hash", "TEXT"},
		{"bio", "TEXT"},
		{"location", "TEXT"},
		{"website", "TEXT"},
		{"avatar_url", "TEXT"},
		{"role", "TEXT DEFAULT 'member'"},
		{"updated_at", "DATETIME"},
	}

	for _, column := range columns {
		if hasColumn, err := userTableHasColumn(db, column.name); err != nil {
			return err
		} else if !hasColumn {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE users ADD COLUMN %s %s", column.name, column.definition)); err != nil {
				return fmt.Errorf("failed to add users.%s: %w", column.name, err)
			}
		}
	}

	if _, err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE username IS NOT NULL AND username != ''"); err != nil {
		return fmt.Errorf("failed to create users username index: %w", err)
	}

	return nil
}

func userTableHasColumn(db *sql.DB, columnName string) (bool, error) {
	return tableHasColumn(db, "users", columnName)
}

func migrateImagesTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"user_id", "INTEGER"},
		{"user_category", "TEXT"},
		{"captured_at", "TEXT"},
		{"is_public", "INTEGER DEFAULT 1"},
	}

	for _, column := range columns {
		if hasColumn, err := tableHasColumn(db, "images", column.name); err != nil {
			return err
		} else if !hasColumn {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE images ADD COLUMN %s %s", column.name, column.definition)); err != nil {
				return fmt.Errorf("failed to add images.%s: %w", column.name, err)
			}
		}
	}

	return nil
}

func migratePhotosTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"user_id", "INTEGER"},
		{"user_category", "TEXT"},
		{"file_name", "TEXT"},
		{"captured_at", "TEXT"},
		{"photo_location", "TEXT"},
		{"camera", "TEXT"},
		{"focal_length", "TEXT"},
		{"is_public", "INTEGER DEFAULT 1"},
	}

	for _, column := range columns {
		if hasColumn, err := tableHasColumn(db, "photos", column.name); err != nil {
			return err
		} else if !hasColumn {
			if _, err := db.Exec(fmt.Sprintf("ALTER TABLE photos ADD COLUMN %s %s", column.name, column.definition)); err != nil {
				return fmt.Errorf("failed to add photos.%s: %w", column.name, err)
			}
		}
	}

	return nil
}

func copyPhotoMediaFromImages(db *sql.DB) error {
	hasImages, err := tableExists(db, "images")
	if err != nil {
		return err
	}
	hasPhotos, err := tableExists(db, "photos")
	if err != nil {
		return err
	}
	if !hasImages || !hasPhotos {
		return nil
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO photos (
			id, user_id, title, description, image_path, thumbnail, category, user_category,
			file_name, dimensions, file_type, file_size, orientation, resolution, color_mode,
			photographer, captured_at, photo_location, camera, focal_length,
			is_public, created_at, updated_at
		)
		SELECT
			id, user_id, title, description, image_path, thumbnail, 'Photos', user_category,
			'', dimensions, file_type, file_size, orientation, resolution, color_mode,
			photographer,
			COALESCE(captured_at, ''), '', '', '',
			COALESCE(is_public, 1), created_at, updated_at
		FROM images
		WHERE category = 'Photos'
	`); err != nil {
		return fmt.Errorf("failed to copy photo media from images: %w", err)
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO photo_tags (photo_id, tag_id)
		SELECT image_id, tag_id
		FROM image_tags
		WHERE image_id IN (SELECT id FROM photos)
	`); err != nil {
		return fmt.Errorf("failed to copy photo tags from images: %w", err)
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO photo_keywords (photo_id, keyword_id)
		SELECT image_id, keyword_id
		FROM image_keywords
		WHERE image_id IN (SELECT id FROM photos)
	`); err != nil {
		return fmt.Errorf("failed to copy photo keywords from images: %w", err)
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO collection_photos (collection_id, photo_id, added_at)
		SELECT collection_id, image_id, added_at
		FROM collection_images
		WHERE image_id IN (SELECT id FROM photos)
	`); err != nil {
		return fmt.Errorf("failed to copy collection photos from images: %w", err)
	}

	if _, err := db.Exec(`
		INSERT OR IGNORE INTO photo_likes (user_id, photo_id, added_at)
		SELECT user_id, image_id, added_at
		FROM image_likes
		WHERE image_id IN (SELECT id FROM photos)
	`); err != nil {
		return fmt.Errorf("failed to copy photo likes from images: %w", err)
	}

	return nil
}

func dropLegacyMemberSinceColumns(db *sql.DB) error {
	for _, tableName := range []string{"images", "photos"} {
		hasColumn, err := tableHasColumn(db, tableName, "member_since")
		if err != nil {
			return err
		}
		if !hasColumn {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s DROP COLUMN member_since", tableName)); err != nil {
			return fmt.Errorf("failed to drop %s.member_since: %w", tableName, err)
		}
	}
	return nil
}

func dropLegacyImagePhotoMetadataColumns(db *sql.DB) error {
	for _, columnName := range []string{"photo_location", "focal_length"} {
		hasColumn, err := tableHasColumn(db, "images", columnName)
		if err != nil {
			return err
		}
		if !hasColumn {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE images DROP COLUMN %s", columnName)); err != nil {
			return fmt.Errorf("failed to drop images.%s: %w", columnName, err)
		}
	}
	return nil
}

func dropLegacyImagePriceColumn(db *sql.DB) error {
	hasPrice, err := tableHasColumn(db, "images", "price")
	if err != nil {
		return err
	}
	if !hasPrice {
		return nil
	}
	if _, err := db.Exec("ALTER TABLE images DROP COLUMN price"); err != nil {
		return fmt.Errorf("failed to drop images.price: %w", err)
	}
	return nil
}

func migrateLegacyPhotosTable(db *sql.DB) error {
	hasPhotos, err := tableExists(db, "photos")
	if err != nil {
		return err
	}
	if !hasPhotos {
		return nil
	}

	hasImages, err := tableExists(db, "images")
	if err != nil {
		return err
	}
	if hasImages {
		if err := mergeLegacyPhotosIntoImages(db); err != nil {
			return err
		}
		if _, err := db.Exec("DROP TABLE photos"); err != nil {
			return fmt.Errorf("failed to drop legacy photos table: %w", err)
		}
		return nil
	}

	if _, err := db.Exec("ALTER TABLE photos RENAME TO images"); err != nil {
		return fmt.Errorf("failed to rename photos table to images: %w", err)
	}
	return nil
}

func mergeLegacyPhotosIntoImages(db *sql.DB) error {
	_, err := db.Exec(`
		INSERT OR IGNORE INTO images (
			id, user_id, title, description, image_path, thumbnail, category, user_category,
			dimensions, file_type, file_size, orientation, resolution, color_mode,
			photographer, is_public, created_at, updated_at
		)
		SELECT
			id, user_id, title, description, image_path, thumbnail, category, user_category,
			dimensions, file_type, file_size, orientation, resolution, color_mode,
			photographer, COALESCE(is_public, 1), created_at, updated_at
		FROM photos
	`)
	if err != nil {
		return fmt.Errorf("failed to merge legacy photos into images: %w", err)
	}
	return nil
}

func migrateLegacyImageRelationTables(db *sql.DB) error {
	for _, tableName := range []string{"cart_items"} {
		if err := renameColumnIfNeeded(db, tableName, "photo_id", "image_id"); err != nil {
			return err
		}
	}

	if err := migrateLegacyCartItemsToLikes(db); err != nil {
		return err
	}

	return nil
}

func migrateLegacyCartItemsToLikes(db *sql.DB) error {
	hasCart, err := tableExists(db, "cart_items")
	if err != nil {
		return err
	}
	if !hasCart {
		return nil
	}
	hasLikes, err := tableExists(db, "image_likes")
	if err != nil {
		return err
	}
	if !hasLikes {
		if _, err := db.Exec(`
			CREATE TABLE image_likes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				user_id INTEGER NOT NULL,
				image_id INTEGER NOT NULL,
				added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (user_id) REFERENCES users(id),
				FOREIGN KEY (image_id) REFERENCES images(id),
				UNIQUE (user_id, image_id)
			)
		`); err != nil {
			return fmt.Errorf("failed to create image_likes table: %w", err)
		}
	}
	if hasImageID, err := tableHasColumn(db, "cart_items", "image_id"); err != nil {
		return err
	} else if hasImageID {
		if _, err := db.Exec("INSERT OR IGNORE INTO image_likes (user_id, image_id, added_at) SELECT user_id, image_id, added_at FROM cart_items"); err != nil {
			return fmt.Errorf("failed to migrate cart items to likes: %w", err)
		}
	}
	if _, err := db.Exec("DROP TABLE cart_items"); err != nil {
		return fmt.Errorf("failed to drop legacy cart_items table: %w", err)
	}
	return nil
}

func renameTableIfNeeded(db *sql.DB, oldName string, newName string) error {
	hasOld, err := tableExists(db, oldName)
	if err != nil {
		return err
	}
	if !hasOld {
		return nil
	}
	hasNew, err := tableExists(db, newName)
	if err != nil {
		return err
	}
	if hasNew {
		if err := mergeLegacyRelationTable(db, oldName, newName); err != nil {
			return err
		}
		if _, err := db.Exec(fmt.Sprintf("DROP TABLE %s", oldName)); err != nil {
			return fmt.Errorf("failed to drop legacy %s table: %w", oldName, err)
		}
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME TO %s", oldName, newName)); err != nil {
		return fmt.Errorf("failed to rename %s table to %s: %w", oldName, newName, err)
	}
	return nil
}

func mergeLegacyRelationTable(db *sql.DB, oldName string, newName string) error {
	var statement string
	switch {
	case oldName == "photo_tags" && newName == "image_tags":
		statement = "INSERT OR IGNORE INTO image_tags (image_id, tag_id) SELECT photo_id, tag_id FROM photo_tags"
	case oldName == "photo_keywords" && newName == "image_keywords":
		statement = "INSERT OR IGNORE INTO image_keywords (image_id, keyword_id) SELECT photo_id, keyword_id FROM photo_keywords"
	case oldName == "collection_photos" && newName == "collection_images":
		statement = "INSERT OR IGNORE INTO collection_images (collection_id, image_id, added_at) SELECT collection_id, photo_id, added_at FROM collection_photos"
	default:
		return nil
	}

	if _, err := db.Exec(statement); err != nil {
		return fmt.Errorf("failed to merge legacy %s into %s: %w", oldName, newName, err)
	}
	return nil
}

func renameColumnIfNeeded(db *sql.DB, tableName string, oldName string, newName string) error {
	hasTable, err := tableExists(db, tableName)
	if err != nil {
		return err
	}
	if !hasTable {
		return nil
	}
	hasOld, err := tableHasColumn(db, tableName, oldName)
	if err != nil {
		return err
	}
	if !hasOld {
		return nil
	}
	hasNew, err := tableHasColumn(db, tableName, newName)
	if err != nil {
		return err
	}
	if hasNew {
		return nil
	}
	if _, err := db.Exec(fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", tableName, oldName, newName)); err != nil {
		return fmt.Errorf("failed to rename %s.%s to %s: %w", tableName, oldName, newName, err)
	}
	return nil
}

func tableExists(db *sql.DB, tableName string) (bool, error) {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func tableHasColumn(db *sql.DB, tableName string, columnName string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue interface{}
		var pk int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}

// Helper function to insert sample data
func InsertSampleData(db *sql.DB) error {
	// Insert sample categories
	categories := []string{"dress", "evening dress", "fashion", "portraits"}
	for _, category := range categories {
		_, err := db.Exec("INSERT OR IGNORE INTO categories (name) VALUES (?)", category)
		if err != nil {
			return fmt.Errorf("failed to insert category: %w", err)
		}
	}

	// Insert sample tags
	tags := []string{"elegant", "satin", "champagne", "luxury", "formal"}
	for _, tag := range tags {
		_, err := db.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tag)
		if err != nil {
			return fmt.Errorf("failed to insert tag: %w", err)
		}
	}

	return nil
}
