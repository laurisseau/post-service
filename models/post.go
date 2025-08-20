package models

import (
	"time"
	"github.com/google/uuid"
)

// Post model for LinkedIn-style posts
type Post struct {
	ID            uuid.UUID   `json:"id" db:"id"`                           // Primary key
	UserID        uuid.UUID   `json:"user_id" db:"user_id"`                 // Reference to User
	Caption       string      `json:"caption" db:"caption"`                 // Text content
	Images        []string    `json:"images" db:"images"`                   // Array of S3 image URLs
	Videos        []string    `json:"videos" db:"videos"`                   // Array of S3 video URLs
	LikesCount    int64       `json:"likes_count" db:"likes_count"`         // Count of likes
	CommentsCount int64       `json:"comments_count" db:"comments_count"`   // Count of comments
	CreatedAt     time.Time   `json:"created_at" db:"created_at"`           // Post creation time
	UpdatedAt     time.Time   `json:"updated_at" db:"updated_at"`           // Last updated time
	Visibility    string      `json:"visibility" db:"visibility"`           // "public", "connections", "private"
	Tags          []string    `json:"tags" db:"tags"`                       // Optional hashtags, mentions, etc.
}