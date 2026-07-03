package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"photo-library/config"
	"photo-library/database"
	"photo-library/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Create database tables
	if err := database.CreateTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Initialize router
	router := mux.NewRouter()

	registerRoutes(router, "", db)
	registerRoutes(router, "/mmh", db)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func registerRoutes(router *mux.Router, basePath string, db *sql.DB) {
	// Static files
	router.PathPrefix(basePath + "/static/").Handler(http.StripPrefix(basePath+"/static/", http.FileServer(http.Dir("static"))))

	// Page routes
	router.HandleFunc(basePath+"/", handlers.HomeHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/photos", handlers.PhotosPageHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/my-media", handlers.MediaManagerHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/profile/edit", handlers.ProfileEditHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/upload", handlers.UploadHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/admin", redirectHandler(basePath+"/upload")).Methods("GET")
	router.HandleFunc(basePath+"/login", handlers.LoginPageHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/register", handlers.RegisterPageHandler(db)).Methods("GET")
	router.HandleFunc(basePath+"/photo/{id}", handlers.PhotoDetailHandler(db)).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix(basePath + "/api").Subrouter()

	// Photos
	apiRouter.HandleFunc("/photos", handlers.GetPhotosHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/photos", handlers.CreatePhotoHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/admin/photos", handlers.GetAdminPhotosHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/admin/photos/upload", handlers.UploadPhotoHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/admin/categories", handlers.GetAdminCategoriesHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/admin/categories", handlers.CreateAdminCategoryHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/photos/{id}", handlers.UpdatePhotoHandler(db)).Methods("PUT")
	apiRouter.HandleFunc("/photos/{id}", handlers.DeletePhotoHandler(db)).Methods("DELETE")
	apiRouter.HandleFunc("/photos/{id}/collections", handlers.GetPhotoCollectionsHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/photos/{id}/move", handlers.MovePhotoHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/photos/{id}/visibility", handlers.UpdatePhotoVisibilityHandler(db)).Methods("PUT")

	// Users
	apiRouter.HandleFunc("/me", handlers.CurrentUserHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/me", handlers.UpdateCurrentUserHandler(db)).Methods("PUT", "POST")
	apiRouter.HandleFunc("/register", handlers.RegisterUserHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/login", handlers.LoginUserHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/logout", handlers.LogoutUserHandler(db)).Methods("POST")

	// Search & Filter
	apiRouter.HandleFunc("/search", handlers.SearchHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/tags", handlers.GetTagsHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/categories", handlers.GetCategoriesHandler(db)).Methods("GET")

	// Collections
	apiRouter.HandleFunc("/collections", handlers.GetCollectionsHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/collections", handlers.CreateCollectionHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/collections/{id}", handlers.GetCollectionHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/collections/{id}/add", handlers.AddPhotoToCollectionHandler(db)).Methods("POST")

	// Cart
	apiRouter.HandleFunc("/cart", handlers.GetCartHandler(db)).Methods("GET")
	apiRouter.HandleFunc("/cart/add", handlers.AddToCartHandler(db)).Methods("POST")
	apiRouter.HandleFunc("/cart/remove/{id}", handlers.RemoveFromCartHandler(db)).Methods("DELETE")

}

func redirectHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, path, http.StatusMovedPermanently)
	}
}
