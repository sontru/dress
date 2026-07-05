package database

import "testing"

func TestCreateTablesDropsMediaMemberSinceColumns(t *testing.T) {
	db, err := Init(t.TempDir() + "/media_hub.db")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer db.Close()

	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables() error = %v", err)
	}

	for _, tableName := range []string{"images", "photos"} {
		hasColumn, err := tableHasColumn(db, tableName, "member_since")
		if err != nil {
			t.Fatalf("tableHasColumn(%q) error = %v", tableName, err)
		}
		if hasColumn {
			t.Fatalf("%s.member_since should not exist", tableName)
		}
	}
}

func TestCreateTablesMigratesLegacyMediaMemberSinceColumns(t *testing.T) {
	db, err := Init(t.TempDir() + "/legacy_media_hub.db")
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE images (
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
			captured_at TEXT,
			photo_location TEXT,
			focal_length TEXT,
			is_public INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE photos (
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
			captured_at TEXT,
			photo_location TEXT,
			focal_length TEXT,
			is_public INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("creating legacy tables error = %v", err)
	}

	if err := CreateTables(db); err != nil {
		t.Fatalf("CreateTables() error = %v", err)
	}

	for _, tableName := range []string{"images", "photos"} {
		hasColumn, err := tableHasColumn(db, tableName, "member_since")
		if err != nil {
			t.Fatalf("tableHasColumn(%q) error = %v", tableName, err)
		}
		if hasColumn {
			t.Fatalf("%s.member_since should be dropped", tableName)
		}
	}
}
