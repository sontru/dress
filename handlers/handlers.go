package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"photo-library/models"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// HomeHandler renders the home page
func HomeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/index.html")
	}
}

// ImagesPageHandler renders the image browsing page
func ImagesPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/images.html")
	}
}

// PhotosPageHandler renders the photo browsing page
func PhotosPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/photos.html")
	}
}

// MediaManagerHandler renders the logged-in user media manager page
func MediaManagerHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireCurrentUser(db, w, r); !ok {
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/media-manager.html")
	}
}

// ProfileEditHandler renders the logged-in user profile editor.
func ProfileEditHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireCurrentUser(db, w, r); !ok {
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/profile-edit.html")
	}
}

// UploadHandler renders the logged-in user media upload page.
func UploadHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireCurrentUser(db, w, r); !ok {
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/upload.html")
	}
}

// LoginPageHandler renders the login page
func LoginPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/login.html")
	}
}

// RegisterPageHandler renders the registration page
func RegisterPageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/register.html")
	}
}

// PhotoDetailHandler renders the photo detail page
func PhotoDetailHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		http.ServeFile(w, r, "templates/photo-detail.html")
	}
}

// GetPhotosHandler retrieves all photos with pagination or a single photo by ID
func GetPhotosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if requesting a specific photo by ID
		photoID := r.URL.Query().Get("id")
		if photoID != "" {
			// Get single photo with tags
			var photo models.Photo
			var userCreatedAt string
			err := db.QueryRow(`
				SELECT p.id, COALESCE(p.user_id, 0), p.title, p.description, p.image_path, p.thumbnail, p.category,
					   COALESCE(p.user_category, ''),
					   p.dimensions, p.file_type, p.file_size, p.orientation, p.resolution, 
					   p.color_mode, p.photographer, COALESCE(u.username, ''), COALESCE(u.created_at, ''),
					   COALESCE(p.captured_at, ''), COALESCE(p.photo_location, ''), COALESCE(p.focal_length, ''),
					   COALESCE(p.is_public, 1), p.created_at
				FROM photos p
				LEFT JOIN users u ON p.user_id = u.id
				WHERE p.id = ?
			`, photoID).Scan(&photo.ID, &photo.UserID, &photo.Title, &photo.Description, &photo.ImagePath,
				&photo.Thumbnail, &photo.Category, &photo.UserCategory, &photo.Dimensions, &photo.FileType,
				&photo.FileSize, &photo.Orientation, &photo.Resolution, &photo.ColorMode,
				&photo.Photographer, &photo.PhotographerUsername, &userCreatedAt,
				&photo.CapturedAt, &photo.PhotoLocation, &photo.FocalLength,
				&photo.IsPublic, &photo.CreatedAt)

			if err != nil {
				http.Error(w, "Photo not found", http.StatusNotFound)
				return
			}
			if !photo.IsPublic && !requestUserOwnsPhoto(db, r, photo.ID) {
				http.Error(w, "Photo not found", http.StatusNotFound)
				return
			}
			setPhotoMemberSince(&photo, userCreatedAt)

			// Load tags for this photo
			tagRows, err := db.Query(`
				SELECT t.name FROM tags t
				JOIN photo_tags pt ON t.id = pt.tag_id
				WHERE pt.photo_id = ?
				ORDER BY t.name
			`, photo.ID)

			if err == nil {
				defer tagRows.Close()
				photo.Tags = []string{}
				for tagRows.Next() {
					var tagName string
					if err := tagRows.Scan(&tagName); err == nil {
						photo.Tags = append(photo.Tags, tagName)
					}
				}
			}
			populateLikeState(db, r, &photo)

			// Return single photo as array
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]models.Photo{photo})
			return
		}

		// Original pagination logic
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		pageNum, _ := strconv.Atoi(page)
		pageSize := 12
		pageSizeParam := r.URL.Query().Get("pageSize")
		category := strings.TrimSpace(r.URL.Query().Get("category"))
		if pageSizeParam == "all" {
			pageSize = 0
		} else if parsedPageSize, err := strconv.Atoi(pageSizeParam); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		}
		offset := (pageNum - 1) * pageSize

		query := `
			SELECT p.id, p.title, p.description, p.image_path, p.thumbnail, p.category,
				   p.dimensions, p.file_type, p.file_size, p.orientation, p.resolution,
				   p.color_mode, p.photographer, COALESCE(u.created_at, ''),
				   COALESCE(p.captured_at, ''), COALESCE(p.photo_location, ''), COALESCE(p.focal_length, ''),
				   COALESCE(p.is_public, 1), p.created_at
			FROM photos p
			LEFT JOIN users u ON p.user_id = u.id
		`
		args := []interface{}{}
		if user, err := currentUserFromRequest(db, r); err == nil {
			query += ` WHERE (COALESCE(p.is_public, 1) = 1 OR p.user_id = ?)`
			args = append(args, user.ID)
		} else {
			query += ` WHERE COALESCE(p.is_public, 1) = 1`
		}
		if category != "" {
			query += ` AND p.category = ?`
			args = append(args, category)
		}
		query += ` ORDER BY p.created_at DESC, p.id DESC`
		if pageSize > 0 {
			query += ` LIMIT ? OFFSET ?`
			args = append(args, pageSize, offset)
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		photos := []models.Photo{}
		for rows.Next() {
			var photo models.Photo
			var userCreatedAt string
			err := rows.Scan(&photo.ID, &photo.Title, &photo.Description, &photo.ImagePath,
				&photo.Thumbnail, &photo.Category, &photo.Dimensions, &photo.FileType,
				&photo.FileSize, &photo.Orientation, &photo.Resolution, &photo.ColorMode,
				&photo.Photographer, &userCreatedAt, &photo.CapturedAt, &photo.PhotoLocation,
				&photo.FocalLength, &photo.IsPublic, &photo.CreatedAt)
			if err != nil {
				continue
			}
			setPhotoMemberSince(&photo, userCreatedAt)
			populateLikeState(db, r, &photo)
			photos = append(photos, photo)
		}

		// Get total count
		var total int
		countQuery := "SELECT COUNT(*) FROM photos WHERE COALESCE(is_public, 1) = 1"
		countArgs := []interface{}{}
		if user, err := currentUserFromRequest(db, r); err == nil {
			countQuery = "SELECT COUNT(*) FROM photos WHERE (COALESCE(is_public, 1) = 1 OR user_id = ?)"
			countArgs = append(countArgs, user.ID)
		}
		if category != "" {
			countQuery += " AND category = ?"
			countArgs = append(countArgs, category)
		}
		db.QueryRow(countQuery, countArgs...).Scan(&total)

		result := models.SearchResult{
			Photos:   photos,
			Total:    total,
			Page:     pageNum,
			PageSize: pageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// CreatePhotoHandler creates a new photo
func CreatePhotoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var photo models.Photo
		if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec(`
			INSERT INTO photos (title, description, image_path, thumbnail, category,
				file_name, dimensions, file_type, file_size, orientation, resolution,
				color_mode, photographer, captured_at, photo_location, camera, focal_length)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, photo.Title, photo.Description, photo.ImagePath, photo.Thumbnail,
			defaultString(photo.Category, "Photos"), photo.FileName, photo.Dimensions, photo.FileType, photo.FileSize,
			photo.Orientation, photo.Resolution, photo.ColorMode,
			photo.Photographer, photo.CapturedAt, photo.PhotoLocation, photo.Camera,
			photo.FocalLength)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}

// RegisterUserHandler creates a new user profile
func RegisterUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerUserRequest
		if err := decodeFormOrJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req.normalize()
		if err := req.validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		passwordHash, err := hashPassword(req.Password)
		if err != nil {
			http.Error(w, "Unable to secure password", http.StatusInternalServerError)
			return
		}

		result, err := db.Exec(`
			INSERT INTO users (username, email, name, real_name, password_hash, bio, location, website, avatar_url, role)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, req.Username, req.Email, req.Name, req.RealName, passwordHash, req.Bio, req.Location, req.Website, req.AvatarURL, "member")
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "unique") {
				http.Error(w, "Username or email is already registered", http.StatusConflict)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := models.User{
			ID:        int(id),
			Username:  req.Username,
			Email:     req.Email,
			Name:      req.Name,
			RealName:  req.RealName,
			Bio:       req.Bio,
			Location:  req.Location,
			Website:   req.Website,
			AvatarURL: req.AvatarURL,
			Role:      "member",
		}

		if err := ensureDefaultUserCategories(db, user.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setUserCookie(w, user.ID)
		writeAuthResponse(w, r, http.StatusCreated, user, "/")
	}
}

// LoginUserHandler authenticates an existing user
func LoginUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginUserRequest
		if err := decodeFormOrJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req.Login = strings.TrimSpace(req.Login)
		if req.Login == "" || req.Password == "" {
			http.Error(w, "Email, username, and password are required", http.StatusBadRequest)
			return
		}

		var user models.User
		var storedHash string
		err := db.QueryRow(`
			SELECT id,
				COALESCE(username, ''),
				COALESCE(email, ''),
				COALESCE(name, ''),
				COALESCE(real_name, ''),
				COALESCE(bio, ''),
				COALESCE(location, ''),
				COALESCE(website, ''),
				COALESCE(avatar_url, ''),
				COALESCE(role, 'member'),
				COALESCE(password_hash, '')
			FROM users
			WHERE email = ? OR username = ?
		`, req.Login, req.Login).Scan(&user.ID, &user.Username, &user.Email, &user.Name, &user.RealName,
			&user.Bio, &user.Location, &user.Website, &user.AvatarURL, &user.Role, &storedHash)
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid login details", http.StatusUnauthorized)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !checkPassword(req.Password, storedHash) {
			http.Error(w, "Invalid login details", http.StatusUnauthorized)
			return
		}

		setUserCookie(w, user.ID)
		writeAuthResponse(w, r, http.StatusOK, user, "/")
	}
}

// CurrentUserHandler returns the logged-in user profile
func CurrentUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

// UpdateCurrentUserHandler updates the logged-in user profile and optional avatar image.
func UpdateCurrentUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		if err := r.ParseMultipartForm(8 << 20); err != nil {
			http.Error(w, "Profile update must be 8MB or smaller", http.StatusBadRequest)
			return
		}

		updated := user
		updated.Name = strings.TrimSpace(r.FormValue("name"))
		updated.RealName = strings.TrimSpace(r.FormValue("real_name"))
		updated.Bio = strings.TrimSpace(r.FormValue("bio"))
		updated.Location = strings.TrimSpace(r.FormValue("location"))
		updated.Website = strings.TrimSpace(r.FormValue("website"))
		if avatarURL := strings.TrimSpace(r.FormValue("avatar_url")); avatarURL != "" {
			updated.AvatarURL = avatarURL
		}

		if updated.Name == "" {
			http.Error(w, "Display name is required", http.StatusBadRequest)
			return
		}
		if updated.RealName == "" {
			http.Error(w, "Real name is required", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			avatarPath, err := saveAvatarUpload(user, file, header)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			updated.AvatarURL = avatarPath
		} else if err != http.ErrMissingFile {
			http.Error(w, "Unable to read avatar upload", http.StatusBadRequest)
			return
		}

		_, err = db.Exec(`
			UPDATE users
			SET name = ?, real_name = ?, bio = ?, location = ?, website = ?, avatar_url = ?, updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, updated.Name, updated.RealName, updated.Bio, updated.Location, updated.Website, updated.AvatarURL, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updated)
	}
}

// LogoutUserHandler clears the current user session cookie.
func LogoutUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clearUserCookie(w)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Signed out"})
	}
}

// GetAdminCategoriesHandler returns categories owned by the logged-in user
func GetAdminCategoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		if err := ensureDefaultUserCategories(db, user.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		rows, err := db.Query(`
			SELECT id, user_id, name
			FROM user_categories
			WHERE user_id = ?
			ORDER BY name
		`, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		categories := []models.Category{}
		for rows.Next() {
			var category models.Category
			if err := rows.Scan(&category.ID, &category.UserID, &category.Name); err != nil {
				continue
			}
			categories = append(categories, category)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}
}

// CreateAdminCategoryHandler creates a category owned by the logged-in user.
func CreateAdminCategoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			http.Error(w, "Category name is required", http.StatusBadRequest)
			return
		}

		result, err := db.Exec(`
			INSERT OR IGNORE INTO user_categories (user_id, name, updated_at)
			VALUES (?, ?, CURRENT_TIMESTAMP)
		`, user.ID, req.Name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.Category{
			ID:     int(id),
			UserID: user.ID,
			Name:   req.Name,
		})
	}
}

// GetAdminPhotosHandler returns photos owned by the logged-in user
func GetAdminPhotosHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		rows, err := db.Query(`
			SELECT id, COALESCE(user_id, 0), title, description, image_path, thumbnail, category,
				   COALESCE(user_category, ''),
				   COALESCE(file_name, ''), dimensions, file_type, file_size, orientation, resolution,
				   color_mode, photographer, ?,
				   COALESCE(captured_at, ''), COALESCE(photo_location, ''), COALESCE(camera, ''), COALESCE(focal_length, ''),
				   COALESCE(is_public, 1), created_at
			FROM photos
			WHERE user_id = ?
			UNION ALL
			SELECT id, COALESCE(user_id, 0), title, description, image_path, thumbnail, category,
				   COALESCE(user_category, ''),
				   '', dimensions, file_type, file_size, orientation, resolution,
				   color_mode, photographer, ?,
				   COALESCE(captured_at, ''), '', '', '',
				   COALESCE(is_public, 1), created_at
			FROM images
			WHERE user_id = ?
			ORDER BY created_at DESC, id DESC
		`, user.MemberSince, user.ID, user.MemberSince, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		photos := []models.Photo{}
		for rows.Next() {
			var photo models.Photo
			err := rows.Scan(&photo.ID, &photo.UserID, &photo.Title, &photo.Description, &photo.ImagePath,
				&photo.Thumbnail, &photo.Category, &photo.UserCategory, &photo.FileName, &photo.Dimensions, &photo.FileType,
				&photo.FileSize, &photo.Orientation, &photo.Resolution, &photo.ColorMode,
				&photo.Photographer, &photo.MemberSince, &photo.CapturedAt, &photo.PhotoLocation,
				&photo.Camera, &photo.FocalLength, &photo.IsPublic, &photo.CreatedAt)
			if err != nil {
				continue
			}
			populateLikeState(db, r, &photo)
			photos = append(photos, photo)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(photos)
	}
}

// UploadPhotoHandler uploads a photo, extracts image metadata, and creates a photo record
func UploadPhotoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, "Upload must be 20MB or smaller", http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("photo")
		if err != nil {
			http.Error(w, "Photo file is required", http.StatusBadRequest)
			return
		}
		defer file.Close()

		photo, imageBytes, err := buildUploadedPhoto(r, file, header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userUploadDir := userStorageFolder(user)
		uploadDir := mediaUploadDir(user, photo.Category, photo.UserCategory)
		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			http.Error(w, "Unable to prepare upload directory", http.StatusInternalServerError)
			return
		}

		filename := uniqueImageFilename(header.Filename, photo.FileType)
		diskPath := filepath.Join(uploadDir, filename)
		if err := os.WriteFile(diskPath, imageBytes, 0644); err != nil {
			http.Error(w, "Unable to save uploaded photo", http.StatusInternalServerError)
			return
		}

		if strings.EqualFold(r.FormValue("apply_watermark"), "true") {
			watermarkFile, watermarkHeader, err := r.FormFile("watermark")
			if err != nil {
				http.Error(w, "Watermark file is required when watermarking is enabled", http.StatusBadRequest)
				return
			}
			defer watermarkFile.Close()

			if _, err := saveWatermarkUpload(user, watermarkFile, watermarkHeader); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		photo.UserID = user.ID
		photo.MemberSince = user.MemberSince
		photo.FileName = header.Filename
		photo.IsPublic = true
		photo.ImagePath = mediaPublicPath(userUploadDir, photo.Category, photo.UserCategory, filename)
		photo.Thumbnail = photo.ImagePath

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		var result sql.Result
		if isPhotoCategory(photo.Category) {
			result, err = insertUploadedPhoto(tx, photo)
		} else {
			result, err = insertUploadedImage(tx, photo)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, err := result.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		photo.ID = int(id)

		if err := attachMediaTags(tx, photo.ID, parseCommaList(r.FormValue("tags")), isPhotoCategory(photo.Category)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := upsertUserCategory(tx, user.ID, photo.UserCategory); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(photo)
	}
}

// UpdatePhotoHandler updates an existing photo
func UpdatePhotoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		id, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}

		var photo models.Photo
		if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var current storedMediaPaths
		tableName, err := userMediaTable(db, user.ID, id)
		if err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}
		if err := db.QueryRow(fmt.Sprintf(`
			SELECT image_path, thumbnail, category, COALESCE(user_category, '')
			FROM %s
			WHERE id = ? AND user_id = ?
		`, tableName), id, user.ID).Scan(&current.ImagePath, &current.Thumbnail, &current.Category, &current.UserCategory); err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		moved, err := moveStoredMediaFiles(user, current, photo.Category, photo.UserCategory)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var result sql.Result
		if tableName == "photos" {
			result, err = db.Exec(`
				UPDATE photos SET title=?, description=?, image_path=?, thumbnail=?, category=?, user_category=?, captured_at=?, photo_location=?, camera=?, focal_length=?, is_public=?, updated_at=CURRENT_TIMESTAMP
				WHERE id=? AND user_id=?
			`, photo.Title, photo.Description, moved.ImagePath, moved.Thumbnail, photo.Category, photo.UserCategory,
				photo.CapturedAt, photo.PhotoLocation, photo.Camera, photo.FocalLength, photo.IsPublic, id, user.ID)
		} else {
			result, err = db.Exec(`
				UPDATE images SET title=?, description=?, image_path=?, thumbnail=?, category=?, user_category=?, captured_at=?, is_public=?, updated_at=CURRENT_TIMESTAMP
				WHERE id=? AND user_id=?
			`, photo.Title, photo.Description, moved.ImagePath, moved.Thumbnail, photo.Category, photo.UserCategory,
				photo.CapturedAt, photo.IsPublic, id, user.ID)
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo updated successfully"})
	}
}

// DeletePhotoHandler deletes a photo
func DeletePhotoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		id, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		tableName, err := userMediaTable(db, user.ID, id)
		if err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}
		var result sql.Result
		if tableName == "photos" {
			if _, err := tx.Exec("DELETE FROM collection_photos WHERE photo_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM photo_tags WHERE photo_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM photo_keywords WHERE photo_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM photo_likes WHERE photo_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result, err = tx.Exec("DELETE FROM photos WHERE id=? AND user_id=?", id, user.ID)
		} else {
			if _, err := tx.Exec("DELETE FROM collection_images WHERE image_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM image_tags WHERE image_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM image_keywords WHERE image_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("DELETE FROM image_likes WHERE image_id = ?", id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			result, err = tx.Exec("DELETE FROM images WHERE id=? AND user_id=?", id, user.ID)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo deleted successfully"})
	}
}

// SearchHandler searches photos by tags, keywords, and description
func SearchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		category := r.URL.Query().Get("category")
		tags := r.URL.Query().Get("tags")
		page := r.URL.Query().Get("page")

		if page == "" {
			page = "1"
		}
		pageNum, _ := strconv.Atoi(page)
		pageSize := 12
		offset := (pageNum - 1) * pageSize

		sqlQuery := `
			SELECT DISTINCT p.id, p.title, p.description, p.image_path, p.thumbnail,
				   p.category, p.dimensions, p.file_type, p.file_size, p.orientation,
				   p.resolution, p.color_mode, p.photographer, COALESCE(u.created_at, ''),
				   COALESCE(p.captured_at, ''), COALESCE(p.photo_location, ''), COALESCE(p.focal_length, ''),
				   COALESCE(p.is_public, 1), p.created_at
			FROM photos p
			LEFT JOIN users u ON p.user_id = u.id
			LEFT JOIN photo_tags pt ON p.id = pt.photo_id
			LEFT JOIN tags t ON pt.tag_id = t.id
			WHERE 1=1
		`
		args := []interface{}{}
		countQuery := `
			SELECT COUNT(DISTINCT p.id) FROM photos p
			LEFT JOIN photo_tags pt ON p.id = pt.photo_id
			LEFT JOIN tags t ON pt.tag_id = t.id
			WHERE 1=1
		`
		countArgs := []interface{}{}
		if user, err := currentUserFromRequest(db, r); err == nil {
			sqlQuery += ` AND (COALESCE(p.is_public, 1) = 1 OR p.user_id = ?)`
			countQuery += ` AND (COALESCE(p.is_public, 1) = 1 OR p.user_id = ?)`
			args = append(args, user.ID)
			countArgs = append(countArgs, user.ID)
		} else {
			sqlQuery += ` AND COALESCE(p.is_public, 1) = 1`
			countQuery += ` AND COALESCE(p.is_public, 1) = 1`
		}

		if query != "" {
			sqlQuery += ` AND (p.title LIKE ? OR p.description LIKE ? OR t.name LIKE ?)`
			countQuery += ` AND (p.title LIKE ? OR p.description LIKE ? OR t.name LIKE ?)`
			searchTerm := "%" + query + "%"
			args = append(args, searchTerm, searchTerm, searchTerm)
			countArgs = append(countArgs, searchTerm, searchTerm, searchTerm)
		}

		if category != "" {
			sqlQuery += ` AND p.category = ?`
			countQuery += ` AND p.category = ?`
			args = append(args, category)
			countArgs = append(countArgs, category)
		}

		if tags != "" {
			tagList := strings.Split(tags, ",")
			sqlQuery += ` AND t.name IN (` + strings.Repeat("?,", len(tagList)-1) + `?)`
			countQuery += ` AND t.name IN (` + strings.Repeat("?,", len(tagList)-1) + `?)`
			for _, tag := range tagList {
				tag = strings.TrimSpace(tag)
				args = append(args, tag)
				countArgs = append(countArgs, tag)
			}
		}

		sqlQuery += ` ORDER BY p.created_at DESC, p.id DESC LIMIT ? OFFSET ?`
		args = append(args, pageSize, offset)

		rows, err := db.Query(sqlQuery, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		photos := []models.Photo{}
		for rows.Next() {
			var photo models.Photo
			var userCreatedAt string
			err := rows.Scan(&photo.ID, &photo.Title, &photo.Description, &photo.ImagePath,
				&photo.Thumbnail, &photo.Category, &photo.Dimensions, &photo.FileType,
				&photo.FileSize, &photo.Orientation, &photo.Resolution, &photo.ColorMode,
				&photo.Photographer, &userCreatedAt, &photo.CapturedAt, &photo.PhotoLocation,
				&photo.FocalLength, &photo.IsPublic, &photo.CreatedAt)
			if err != nil {
				continue
			}
			setPhotoMemberSince(&photo, userCreatedAt)
			populateLikeState(db, r, &photo)
			photos = append(photos, photo)
		}

		// Get total count
		var total int
		if err := db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result := models.SearchResult{
			Photos:   photos,
			Total:    total,
			Page:     pageNum,
			PageSize: pageSize,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

// GetTagsHandler retrieves all tags
func GetTagsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name FROM tags ORDER BY name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		tags := []models.Tag{}
		for rows.Next() {
			var tag models.Tag
			rows.Scan(&tag.ID, &tag.Name)
			tags = append(tags, tag)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tags)
	}
}

// GetCategoriesHandler retrieves all categories
func GetCategoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, icon FROM categories ORDER BY name")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		categories := []models.Category{}
		for rows.Next() {
			var cat models.Category
			rows.Scan(&cat.ID, &cat.Name, &cat.Icon)
			categories = append(categories, cat)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	}
}

// GetCollectionsHandler retrieves all collections
func GetCollectionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.URL.Query().Get("user_id")

		query := `
			SELECT id, name, description, user_id, 
				   (SELECT COUNT(*) FROM collection_images WHERE collection_id = collections.id) as photo_count,
				   created_at
			FROM collections
		`
		args := []interface{}{}

		if userID != "" {
			query += " WHERE user_id = ?"
			args = append(args, userID)
		}

		query += " ORDER BY created_at DESC"

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		collections := []models.Collection{}
		for rows.Next() {
			var col models.Collection
			rows.Scan(&col.ID, &col.Name, &col.Description, &col.UserID, &col.PhotoCount, &col.CreatedAt)
			collections = append(collections, col)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}
}

// CreateCollectionHandler creates a new collection
func CreateCollectionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var col models.Collection
		if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := db.Exec(
			"INSERT INTO collections (name, description, user_id) VALUES (?, ?, ?)",
			col.Name, col.Description, col.UserID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int64{"id": id})
	}
}

// GetCollectionHandler retrieves a single collection with photos
func GetCollectionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid collection ID", http.StatusBadRequest)
			return
		}

		// Get collection
		var col models.Collection
		row := db.QueryRow(
			`SELECT id, name, description, user_id, created_at FROM collections WHERE id = ?`,
			id)
		err = row.Scan(&col.ID, &col.Name, &col.Description, &col.UserID, &col.CreatedAt)
		if err != nil {
			http.Error(w, "Collection not found", http.StatusNotFound)
			return
		}

		// Get photos in collection
		rows, err := db.Query(`
			SELECT p.id, p.title, p.image_path, p.thumbnail
			FROM images p
			JOIN collection_images cp ON p.id = cp.image_id
			WHERE cp.collection_id = ?
		`, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		photos := []models.Photo{}
		for rows.Next() {
			var photo models.Photo
			rows.Scan(&photo.ID, &photo.Title, &photo.ImagePath, &photo.Thumbnail)
			populateLikeState(db, r, &photo)
			photos = append(photos, photo)
		}

		col.PhotoCount = len(photos)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"collection": col,
			"photos":     photos,
		})
	}
}

// AddPhotoToCollectionHandler adds a photo to a collection
func AddPhotoToCollectionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		collectionID := mux.Vars(r)["id"]
		if collectionID == "" {
			collectionID = r.URL.Query().Get("id")
		}
		var req struct {
			PhotoID int `json:"photo_id"`
			ImageID int `json:"image_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.ImageID != 0 {
			req.PhotoID = req.ImageID
		}

		_, err := db.Exec(
			"INSERT OR IGNORE INTO collection_images (collection_id, image_id) VALUES (?, ?)",
			collectionID, req.PhotoID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo added to collection"})
	}
}

// GetPhotoCollectionsHandler returns collections containing a photo for its owner.
func GetPhotoCollectionsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		photoID, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}
		if !userOwnsPhoto(db, user.ID, photoID) {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		rows, err := db.Query(`
			SELECT c.id, c.name, c.description, c.user_id,
				   (SELECT COUNT(*) FROM collection_images WHERE collection_id = c.id) AS photo_count,
				   c.created_at
			FROM collections c
			JOIN collection_images cp ON cp.collection_id = c.id
			WHERE cp.image_id = ? AND c.user_id = ?
			ORDER BY c.name
		`, photoID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		collections := []models.Collection{}
		for rows.Next() {
			var col models.Collection
			if err := rows.Scan(&col.ID, &col.Name, &col.Description, &col.UserID, &col.PhotoCount, &col.CreatedAt); err == nil {
				collections = append(collections, col)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(collections)
	}
}

// MovePhotoHandler moves an owned photo into one selected collection.
func MovePhotoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		photoID, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}

		var req struct {
			CollectionID int `json:"collection_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.CollectionID <= 0 {
			http.Error(w, "Collection is required", http.StatusBadRequest)
			return
		}
		if !userOwnsPhoto(db, user.ID, photoID) {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		tableName, err := userMediaTable(db, user.ID, photoID)
		if err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}
		var current storedMediaPaths
		if err := db.QueryRow(fmt.Sprintf(`
			SELECT image_path, thumbnail, category, COALESCE(user_category, '')
			FROM %s
			WHERE id = ? AND user_id = ?
		`, tableName), photoID, user.ID).Scan(&current.ImagePath, &current.Thumbnail, &current.Category, &current.UserCategory); err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		var collectionName string
		if err := db.QueryRow("SELECT name FROM collections WHERE id = ? AND user_id = ?", req.CollectionID, user.ID).Scan(&collectionName); err != nil {
			http.Error(w, "Collection not found for this user", http.StatusNotFound)
			return
		}
		moved, err := moveStoredMediaFiles(user, current, current.Category, collectionName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		if tableName == "photos" {
			if _, err := tx.Exec("DELETE FROM collection_photos WHERE photo_id = ?", photoID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("INSERT OR IGNORE INTO collection_photos (collection_id, photo_id) VALUES (?, ?)", req.CollectionID, photoID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("UPDATE photos SET image_path = ?, thumbnail = ?, user_category = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?", moved.ImagePath, moved.Thumbnail, collectionName, photoID, user.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			if _, err := tx.Exec("DELETE FROM collection_images WHERE image_id = ?", photoID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("INSERT OR IGNORE INTO collection_images (collection_id, image_id) VALUES (?, ?)", req.CollectionID, photoID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if _, err := tx.Exec("UPDATE images SET image_path = ?, thumbnail = ?, user_category = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?", moved.ImagePath, moved.Thumbnail, collectionName, photoID, user.ID); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		if err := upsertUserCategory(tx, user.ID, collectionName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Photo moved"})
	}
}

// UpdatePhotoVisibilityHandler makes an owned photo public or private.
func UpdatePhotoVisibilityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}

		photoID, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid photo ID", http.StatusBadRequest)
			return
		}

		var req struct {
			IsPublic bool `json:"is_public"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tableName, err := userMediaTable(db, user.ID, photoID)
		if err != nil {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}
		result, err := db.Exec(fmt.Sprintf("UPDATE %s SET is_public = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?", tableName), req.IsPublic, photoID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
			http.Error(w, "Photo not found for this user", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Visibility updated"})
	}
}

// LikeImageHandler records a single like from the current user for an image.
func LikeImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}
		imageID, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid image ID", http.StatusBadRequest)
			return
		}
		if !imageIsVisibleToUser(db, user.ID, imageID) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		if _, err := db.Exec("INSERT OR IGNORE INTO image_likes (user_id, image_id) VALUES (?, ?)", user.ID, imageID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeLikeState(w, db, user.ID, imageID)
	}
}

// UnlikeImageHandler removes the current user's like from an image.
func UnlikeImageHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := requireCurrentUser(db, w, r)
		if !ok {
			return
		}
		imageID, err := photoIDFromRequest(r)
		if err != nil {
			http.Error(w, "Invalid image ID", http.StatusBadRequest)
			return
		}
		if !imageIsVisibleToUser(db, user.ID, imageID) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		if _, err := db.Exec("DELETE FROM image_likes WHERE user_id = ? AND image_id = ?", user.ID, imageID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeLikeState(w, db, user.ID, imageID)
	}
}

func buildUploadedPhoto(r *http.Request, file multipart.File, header *multipart.FileHeader) (models.Photo, []byte, error) {
	imageBytes, err := io.ReadAll(io.LimitReader(file, 20<<20+1))
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("unable to read uploaded photo")
	}
	if len(imageBytes) == 0 {
		return models.Photo{}, nil, fmt.Errorf("uploaded photo is empty")
	}
	if len(imageBytes) > 20<<20 {
		return models.Photo{}, nil, fmt.Errorf("upload must be 20MB or smaller")
	}

	cfg, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("uploaded file must be a supported image: JPEG, PNG, or GIF")
	}

	decoded, _, err := image.Decode(bytes.NewReader(imageBytes))
	if err != nil {
		return models.Photo{}, nil, fmt.Errorf("unable to decode uploaded image")
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		return models.Photo{}, nil, fmt.Errorf("title is required")
	}

	photo := models.Photo{
		Title:         title,
		Description:   strings.TrimSpace(r.FormValue("description")),
		Category:      defaultString(strings.TrimSpace(r.FormValue("category")), "Default"),
		UserCategory:  defaultString(strings.TrimSpace(r.FormValue("user_category")), "Default"),
		Dimensions:    fmt.Sprintf("%d x %d px", cfg.Width, cfg.Height),
		FileType:      strings.ToUpper(format),
		FileSize:      humanFileSize(int64(len(imageBytes))),
		Orientation:   imageOrientation(cfg.Width, cfg.Height),
		Resolution:    fmt.Sprintf("%d x %d", cfg.Width, cfg.Height),
		ColorMode:     colorMode(decoded),
		Photographer:  strings.TrimSpace(r.FormValue("photographer")),
		CapturedAt:    strings.TrimSpace(r.FormValue("captured_at")),
		PhotoLocation: strings.TrimSpace(r.FormValue("photo_location")),
		Camera:        strings.TrimSpace(r.FormValue("camera")),
		FocalLength:   strings.TrimSpace(r.FormValue("focal_length")),
		Tags:          parseCommaList(r.FormValue("tags")),
	}

	return photo, imageBytes, nil
}

func saveAvatarUpload(user models.User, file multipart.File, header *multipart.FileHeader) (string, error) {
	imageBytes, err := io.ReadAll(io.LimitReader(file, 8<<20+1))
	if err != nil {
		return "", fmt.Errorf("unable to read avatar upload")
	}
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("avatar upload is empty")
	}
	if len(imageBytes) > 8<<20 {
		return "", fmt.Errorf("avatar must be 8MB or smaller")
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		return "", fmt.Errorf("avatar must be a supported image: JPEG, PNG, or GIF")
	}

	userUploadDir := sanitizeFilename(user.Username)
	if userUploadDir == "" {
		userUploadDir = strconv.Itoa(user.ID)
	}

	uploadDir := filepath.Join("static", "uploads", "avatars", userUploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("unable to prepare avatar upload directory")
	}

	filename := uniqueImageFilename(header.Filename, format)
	diskPath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(diskPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("unable to save avatar upload")
	}

	return "/static/uploads/avatars/" + userUploadDir + "/" + filename, nil
}

func saveWatermarkUpload(user models.User, file multipart.File, header *multipart.FileHeader) (string, error) {
	imageBytes, err := io.ReadAll(io.LimitReader(file, 5<<20+1))
	if err != nil {
		return "", fmt.Errorf("unable to read watermark upload")
	}
	if len(imageBytes) == 0 {
		return "", fmt.Errorf("watermark upload is empty")
	}
	if len(imageBytes) > 5<<20 {
		return "", fmt.Errorf("watermark must be 5MB or smaller")
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(imageBytes))
	if err != nil {
		return "", fmt.Errorf("watermark must be a supported image: JPEG, PNG, or GIF")
	}

	userUploadDir := sanitizeFilename(user.Username)
	if userUploadDir == "" {
		userUploadDir = strconv.Itoa(user.ID)
	}

	uploadDir := filepath.Join("static", "uploads", "watermarks", userUploadDir)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("unable to prepare watermark upload directory")
	}

	filename := uniqueImageFilename(header.Filename, format)
	diskPath := filepath.Join(uploadDir, filename)
	if err := os.WriteFile(diskPath, imageBytes, 0644); err != nil {
		return "", fmt.Errorf("unable to save watermark upload")
	}

	return "/static/uploads/watermarks/" + userUploadDir + "/" + filename, nil
}

type storedMediaPaths struct {
	ImagePath    string
	Thumbnail    string
	Category     string
	UserCategory string
}

func moveStoredMediaFiles(user models.User, current storedMediaPaths, newCategory string, newUserCategory string) (storedMediaPaths, error) {
	moved := current
	targetDir := mediaUploadDir(user, newCategory, newUserCategory)

	imagePath, imageMoved, err := moveStoredMediaFile(current.ImagePath, targetDir)
	if err != nil {
		return moved, err
	}
	if imageMoved {
		moved.ImagePath = imagePath
	}

	if current.Thumbnail == "" || current.Thumbnail == current.ImagePath {
		if imageMoved {
			moved.Thumbnail = imagePath
		}
		moved.Category = newCategory
		moved.UserCategory = newUserCategory
		return moved, nil
	}

	thumbnailPath, thumbnailMoved, err := moveStoredMediaFile(current.Thumbnail, targetDir)
	if err != nil {
		return moved, err
	}
	if thumbnailMoved {
		moved.Thumbnail = thumbnailPath
	}

	moved.Category = newCategory
	moved.UserCategory = newUserCategory
	return moved, nil
}

func moveStoredMediaFile(publicPath string, targetDir string) (string, bool, error) {
	diskPath, ok := publicPathToDiskPath(publicPath)
	if !ok {
		return publicPath, false, nil
	}
	if _, err := os.Stat(diskPath); err != nil {
		if os.IsNotExist(err) {
			return publicPath, false, nil
		}
		return publicPath, false, fmt.Errorf("unable to inspect stored media file")
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return publicPath, false, fmt.Errorf("unable to prepare media category directory")
	}

	targetPath := filepath.Join(targetDir, filepath.Base(diskPath))
	if filepath.Clean(diskPath) == filepath.Clean(targetPath) {
		return publicPath, false, nil
	}
	targetPath = uniqueMoveTarget(targetPath)
	if err := os.Rename(diskPath, targetPath); err != nil {
		return publicPath, false, fmt.Errorf("unable to move stored media file")
	}

	return diskPathToPublicPath(targetPath), true, nil
}

func publicPathToDiskPath(publicPath string) (string, bool) {
	if !strings.HasPrefix(publicPath, "/static/uploads/") {
		return "", false
	}
	cleanPath := filepath.Clean(strings.TrimPrefix(publicPath, "/"))
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") || !strings.HasPrefix(cleanPath, filepath.Join("static", "uploads")+string(filepath.Separator)) {
		return "", false
	}
	return cleanPath, true
}

func diskPathToPublicPath(diskPath string) string {
	return "/" + filepath.ToSlash(filepath.Clean(diskPath))
}

func uniqueMoveTarget(targetPath string) string {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return targetPath
	}

	ext := filepath.Ext(targetPath)
	base := strings.TrimSuffix(targetPath, ext)
	for index := 1; ; index++ {
		candidate := fmt.Sprintf("%s-%d%s", base, index, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func mediaUploadDir(user models.User, mediaCategory string, userCategory string) string {
	return filepath.Join("static", "uploads", userStorageFolder(user), mediaStorageFolder(mediaCategory), categoryStorageFolder(userCategory))
}

func mediaPublicPath(userFolder string, mediaCategory string, userCategory string, filename string) string {
	return "/static/uploads/" + userFolder + "/" + mediaStorageFolder(mediaCategory) + "/" + categoryStorageFolder(userCategory) + "/" + filename
}

func userStorageFolder(user models.User) string {
	folder := sanitizeFilename(user.Username)
	if folder == "" {
		folder = strconv.Itoa(user.ID)
	}
	return folder
}

func mediaStorageFolder(mediaCategory string) string {
	switch strings.ToLower(strings.TrimSpace(mediaCategory)) {
	case "image", "images":
		return "images"
	case "video", "videos":
		return "videos"
	case "audio":
		return "audio"
	default:
		return "photos"
	}
}

func categoryStorageFolder(userCategory string) string {
	folder := sanitizeFilename(userCategory)
	if folder == "" {
		return "default"
	}
	return folder
}

func uniqueImageFilename(originalName string, fileType string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = "." + strings.ToLower(fileType)
	}

	base := strings.TrimSuffix(filepath.Base(originalName), filepath.Ext(originalName))
	base = sanitizeFilename(base)
	if base == "" {
		base = "photo"
	}

	var randomBytes [6]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return fmt.Sprintf("%s-%d%s", base, time.Now().UnixNano(), ext)
	}

	return fmt.Sprintf("%s-%d-%x%s", base, time.Now().UnixNano(), randomBytes, ext)
}

func sanitizeFilename(name string) string {
	name = strings.ToLower(name)
	var builder strings.Builder
	lastDash := false

	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}

	return strings.Trim(builder.String(), "-")
}

func imageOrientation(width int, height int) string {
	switch {
	case width > height:
		return "Landscape"
	case height > width:
		return "Portrait"
	default:
		return "Square"
	}
}

func colorMode(img image.Image) string {
	switch img.(type) {
	case *image.Gray, *image.Gray16:
		return "Grayscale"
	case *image.CMYK:
		return "CMYK"
	case *image.YCbCr:
		return "YCbCr"
	case *image.Paletted:
		return "Indexed color"
	default:
		return "RGB"
	}
}

func humanFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	value := float64(size)
	for _, suffix := range []string{"KB", "MB", "GB"} {
		value /= unit
		if value < unit {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}

	return fmt.Sprintf("%.1f TB", value/unit)
}

func parseCommaList(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	seen := map[string]bool{}

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		key := strings.ToLower(item)
		if seen[key] {
			continue
		}
		seen[key] = true
		items = append(items, item)
	}

	return items
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func displayDate(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	for _, layout := range []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.Format("Jan 2006")
		}
	}

	if len(value) >= 10 {
		return value[:10]
	}
	return value
}

func setPhotoMemberSince(photo *models.Photo, userCreatedAt string) {
	if photo == nil {
		return
	}
	photo.MemberSince = displayDate(userCreatedAt)
}

func insertUploadedPhoto(tx *sql.Tx, photo models.Photo) (sql.Result, error) {
	return tx.Exec(`
		INSERT INTO photos (
			id, user_id, title, description, image_path, thumbnail, category, user_category,
			file_name, dimensions, file_type, file_size, orientation, resolution, color_mode,
			photographer, captured_at, photo_location, camera, focal_length,
			is_public, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, nil, photo.UserID, photo.Title, photo.Description, photo.ImagePath, photo.Thumbnail,
		photo.Category, photo.UserCategory, photo.FileName, photo.Dimensions, photo.FileType,
		photo.FileSize, photo.Orientation, photo.Resolution, photo.ColorMode, photo.Photographer,
		photo.CapturedAt, photo.PhotoLocation, photo.Camera, photo.FocalLength, photo.IsPublic)
}

func insertUploadedImage(tx *sql.Tx, photo models.Photo) (sql.Result, error) {
	return tx.Exec(`
		INSERT INTO images (user_id, title, description, image_path, thumbnail, category, user_category,
			dimensions, file_type, file_size, orientation, resolution,
			color_mode, photographer, captured_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, photo.UserID, photo.Title, photo.Description, photo.ImagePath, photo.Thumbnail,
		photo.Category, photo.UserCategory, photo.Dimensions, photo.FileType, photo.FileSize,
		photo.Orientation, photo.Resolution, photo.ColorMode, photo.Photographer, photo.CapturedAt)
}

func isPhotoCategory(category string) bool {
	return strings.EqualFold(strings.TrimSpace(category), "Photos")
}

func attachTags(tx *sql.Tx, photoID int, tags []string) error {
	return attachMediaTags(tx, photoID, tags, false)
}

func attachMediaTags(tx *sql.Tx, photoID int, tags []string, photoTable bool) error {
	for _, tag := range tags {
		if _, err := tx.Exec("INSERT OR IGNORE INTO tags (name) VALUES (?)", tag); err != nil {
			return err
		}

		var tagID int
		if err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tag).Scan(&tagID); err != nil {
			return err
		}

		if photoTable {
			if _, err := tx.Exec("INSERT OR IGNORE INTO photo_tags (photo_id, tag_id) VALUES (?, ?)", photoID, tagID); err != nil {
				return err
			}
		} else {
			if _, err := tx.Exec("INSERT OR IGNORE INTO image_tags (image_id, tag_id) VALUES (?, ?)", photoID, tagID); err != nil {
				return err
			}
		}
	}

	return nil
}

func upsertUserCategory(tx *sql.Tx, userID int, categoryName string) error {
	categoryName = strings.TrimSpace(categoryName)
	if categoryName == "" {
		categoryName = "Default"
	}

	_, err := tx.Exec(`
		INSERT OR IGNORE INTO user_categories (user_id, name)
		VALUES (?, ?)
	`, userID, categoryName)
	return err
}

func ensureDefaultUserCategories(db *sql.DB, userID int) error {
	defaults := []string{"Default", "Portraits", "Fashion", "Dress", "Events"}
	for _, category := range defaults {
		if _, err := db.Exec(`
			INSERT OR IGNORE INTO user_categories (user_id, name)
			VALUES (?, ?)
		`, userID, category); err != nil {
			return err
		}
	}

	return nil
}

func requireCurrentUser(db *sql.DB, w http.ResponseWriter, r *http.Request) (models.User, bool) {
	user, err := currentUserFromRequest(db, r)
	if err == nil {
		return user, true
	}

	if strings.Contains(r.Header.Get("Accept"), "text/html") {
		http.Redirect(w, r, mountedPath(r, "/login"), http.StatusSeeOther)
	} else {
		http.Error(w, "Login required", http.StatusUnauthorized)
	}

	return models.User{}, false
}

func mountedPath(r *http.Request, path string) string {
	if strings.HasPrefix(r.URL.Path, "/mmh/") || r.URL.Path == "/mmh" {
		return "/mmh" + path
	}
	return path
}

func currentUserFromRequest(db *sql.DB, r *http.Request) (models.User, error) {
	cookie, err := r.Cookie("imagehub_user_id")
	if err != nil {
		return models.User{}, err
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return models.User{}, err
	}

	var user models.User
	var createdAtText string
	err = db.QueryRow(`
		SELECT id,
			COALESCE(username, ''),
			COALESCE(email, ''),
			COALESCE(name, ''),
			COALESCE(real_name, ''),
			COALESCE(bio, ''),
			COALESCE(location, ''),
			COALESCE(website, ''),
			COALESCE(avatar_url, ''),
			COALESCE(role, 'member'),
			COALESCE(created_at, '')
		FROM users
		WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.Name, &user.RealName,
		&user.Bio, &user.Location, &user.Website, &user.AvatarURL, &user.Role, &createdAtText)
	if err != nil {
		return models.User{}, err
	}
	user.MemberSince = displayDate(createdAtText)

	return user, nil
}

func requestUserOwnsPhoto(db *sql.DB, r *http.Request, photoID int) bool {
	user, err := currentUserFromRequest(db, r)
	if err != nil {
		return false
	}
	return userOwnsPhoto(db, user.ID, photoID)
}

func userOwnsPhoto(db *sql.DB, userID int, photoID int) bool {
	_, err := userMediaTable(db, userID, photoID)
	return err == nil
}

func userMediaTable(db *sql.DB, userID int, mediaID int) (string, error) {
	for _, tableName := range []string{"photos", "images"} {
		var exists int
		err := db.QueryRow(fmt.Sprintf("SELECT 1 FROM %s WHERE id = ? AND user_id = ?", tableName), mediaID, userID).Scan(&exists)
		if err == nil {
			return tableName, nil
		}
		if err != sql.ErrNoRows {
			return "", err
		}
	}
	return "", sql.ErrNoRows
}

func photoIDFromRequest(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	if idStr == "" {
		idStr = r.URL.Query().Get("id")
	}
	return strconv.Atoi(idStr)
}

type registerUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	RealName  string `json:"real_name"`
	Password  string `json:"password"`
	Bio       string `json:"bio"`
	Location  string `json:"location"`
	Website   string `json:"website"`
	AvatarURL string `json:"avatar_url"`
}

type loginUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (req *registerUserRequest) normalize() {
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Name = strings.TrimSpace(req.Name)
	req.RealName = strings.TrimSpace(req.RealName)
	req.Bio = strings.TrimSpace(req.Bio)
	req.Location = strings.TrimSpace(req.Location)
	req.Website = strings.TrimSpace(req.Website)
	req.AvatarURL = strings.TrimSpace(req.AvatarURL)
}

func (req registerUserRequest) validate() error {
	switch {
	case req.Username == "":
		return fmt.Errorf("username is required")
	case strings.Contains(req.Username, " "):
		return fmt.Errorf("username cannot contain spaces")
	case req.Email == "":
		return fmt.Errorf("email is required")
	case !strings.Contains(req.Email, "@"):
		return fmt.Errorf("enter a valid email address")
	case req.Name == "":
		return fmt.Errorf("display name is required")
	case req.RealName == "":
		return fmt.Errorf("real name is required")
	case len(req.Password) < 8:
		return fmt.Errorf("password must be at least 8 characters")
	default:
		return nil
	}
}

func decodeFormOrJSON(r *http.Request, target interface{}) error {
	contentType := r.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return json.NewDecoder(r.Body).Decode(target)
	}

	if strings.Contains(contentType, "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			return err
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return err
		}
	}

	switch value := target.(type) {
	case *registerUserRequest:
		value.Username = r.FormValue("username")
		value.Email = r.FormValue("email")
		value.Name = r.FormValue("name")
		value.RealName = r.FormValue("real_name")
		value.Password = r.FormValue("password")
		value.Bio = r.FormValue("bio")
		value.Location = r.FormValue("location")
		value.Website = r.FormValue("website")
		value.AvatarURL = r.FormValue("avatar_url")
	case *loginUserRequest:
		value.Login = r.FormValue("login")
		value.Password = r.FormValue("password")
	default:
		return fmt.Errorf("unsupported request type")
	}

	return nil
}

func hashPassword(password string) (string, error) {
	var salt [16]byte
	if _, err := rand.Read(salt[:]); err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(hex.EncodeToString(salt[:]) + password))
	return hex.EncodeToString(salt[:]) + ":" + hex.EncodeToString(hash[:]), nil
}

func checkPassword(password string, storedHash string) bool {
	parts := strings.Split(storedHash, ":")
	if len(parts) != 2 {
		return false
	}

	hash := sha256.Sum256([]byte(parts[0] + password))
	expected := hex.EncodeToString(hash[:])
	return subtle.ConstantTimeCompare([]byte(expected), []byte(parts[1])) == 1
}

func setUserCookie(w http.ResponseWriter, userID int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "imagehub_user_id",
		Value:    strconv.Itoa(userID),
		Path:     "/",
		MaxAge:   60 * 60 * 24 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearUserCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "imagehub_user_id",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func writeAuthResponse(w http.ResponseWriter, r *http.Request, status int, user models.User, redirectPath string) {
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "text/html") && !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		http.Redirect(w, r, redirectPath, http.StatusSeeOther)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(user)
}

func populateLikeState(db *sql.DB, r *http.Request, photo *models.Photo) {
	if photo == nil || photo.ID == 0 {
		return
	}
	_ = db.QueryRow("SELECT COUNT(*) FROM image_likes WHERE image_id = ?", photo.ID).Scan(&photo.LikeCount)
	if user, err := currentUserFromRequest(db, r); err == nil {
		var liked int
		_ = db.QueryRow("SELECT COUNT(*) FROM image_likes WHERE image_id = ? AND user_id = ?", photo.ID, user.ID).Scan(&liked)
		photo.LikedByUser = liked > 0
	}
}

func writeLikeState(w http.ResponseWriter, db *sql.DB, userID int, imageID int) {
	var likeCount int
	var likedCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM image_likes WHERE image_id = ?", imageID).Scan(&likeCount)
	_ = db.QueryRow("SELECT COUNT(*) FROM image_likes WHERE image_id = ? AND user_id = ?", imageID, userID).Scan(&likedCount)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"image_id":      imageID,
		"like_count":    likeCount,
		"liked_by_user": likedCount > 0,
	})
}

func imageIsVisibleToUser(db *sql.DB, userID int, imageID int) bool {
	var exists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM images
		WHERE id = ? AND (COALESCE(is_public, 1) = 1 OR user_id = ?)
	`, imageID, userID).Scan(&exists)
	return err == nil && exists > 0
}
