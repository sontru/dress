package models

import "time"

type Photo struct {
	ID                   int       `json:"id"`
	UserID               int       `json:"user_id"`
	Title                string    `json:"title"`
	Description          string    `json:"description"`
	ImagePath            string    `json:"image_path"`
	Thumbnail            string    `json:"thumbnail"`
	Category             string    `json:"category"`
	UserCategory         string    `json:"user_category"`
	Tags                 []string  `json:"tags"`
	Keywords             []string  `json:"keywords"`
	Dimensions           string    `json:"dimensions"`
	FileType             string    `json:"file_type"`
	FileSize             string    `json:"file_size"`
	Orientation          string    `json:"orientation"`
	Resolution           string    `json:"resolution"`
	ColorMode            string    `json:"color_mode"`
	Photographer         string    `json:"photographer"`
	PhotographerUsername string    `json:"photographer_username"`
	MemberSince          string    `json:"member_since"`
	Price                float64   `json:"price"`
	IsPublic             bool      `json:"is_public"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type Category struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id,omitempty"`
	Name   string `json:"name"`
	Icon   string `json:"icon"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Collection struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	UserID      int       `json:"user_id"`
	PhotoCount  int       `json:"photo_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItem struct {
	ID       int       `json:"id"`
	PhotoID  int       `json:"photo_id"`
	Quantity int       `json:"quantity"`
	Price    float64   `json:"price"`
	AddedAt  time.Time `json:"added_at"`
}

type User struct {
	ID          int       `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	RealName    string    `json:"real_name"`
	Bio         string    `json:"bio"`
	Location    string    `json:"location"`
	Website     string    `json:"website"`
	AvatarURL   string    `json:"avatar_url"`
	Role        string    `json:"role"`
	MemberSince string    `json:"member_since"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SearchResult struct {
	Photos   []Photo `json:"photos"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
