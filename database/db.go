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

	CREATE TABLE IF NOT EXISTS photos (
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
		member_since TEXT,
		price REAL DEFAULT 29.99,
		is_public INTEGER DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT UNIQUE NOT NULL
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

	CREATE TABLE IF NOT EXISTS collection_photos (
		collection_id INTEGER NOT NULL,
		photo_id INTEGER NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (collection_id) REFERENCES collections(id),
		FOREIGN KEY (photo_id) REFERENCES photos(id),
		PRIMARY KEY (collection_id, photo_id)
	);

	CREATE TABLE IF NOT EXISTS cart_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		photo_id INTEGER NOT NULL,
		quantity INTEGER DEFAULT 1,
		price REAL NOT NULL,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (photo_id) REFERENCES photos(id)
	);

	CREATE INDEX IF NOT EXISTS idx_photos_category ON photos(category);
	CREATE INDEX IF NOT EXISTS idx_photo_tags_tag_id ON photo_tags(tag_id);
	CREATE INDEX IF NOT EXISTS idx_collections_user_id ON collections(user_id);
	CREATE INDEX IF NOT EXISTS idx_cart_items_user_id ON cart_items(user_id);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	if err := migrateUsersTable(db); err != nil {
		return err
	}
	if err := migratePhotosTable(db); err != nil {
		return err
	}
	if err := createPostMigrationIndexes(db); err != nil {
		return err
	}

	return nil
}

func createPostMigrationIndexes(db *sql.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_photos_user_id ON photos(user_id)",
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

func migratePhotosTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"user_id", "INTEGER"},
		{"user_category", "TEXT"},
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

func tableHasColumn(db *sql.DB, tableName string, columnName string) (bool, error) {
	if tableName != "users" && tableName != "photos" {
		return false, fmt.Errorf("unsupported table %q", tableName)
	}

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
