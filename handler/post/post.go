package post

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/laurisseau/post-service/models"
    "github.com/laurisseau/post-service/utils"
	"github.com/lib/pq"
)

func CreateHandler(ctx *gin.Context, db *sql.DB) {
	// Parse multipart form
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	// Read caption
	caption := ""
	if vals, ok := form.Value["caption"]; ok && len(vals) > 0 {
		caption = vals[0]
	}

	// Read visibility
	visibility := "public"
	if vals, ok := form.Value["visibility"]; ok && len(vals) > 0 && vals[0] != "" {
		visibility = vals[0]
	}

	// Read tags
	tags := []string{}
	if vals, ok := form.Value["tags"]; ok && len(vals) > 0 && vals[0] != "" {
		tags = strings.Split(vals[0], ",")
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load AWS config"})
		return
	}

	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

    //get logged-in userid
    UserID := utils.GetProfileIdFromSession(ctx)

	// Build Post model
	post := models.Post{
		ID:            uuid.New(),
		UserID:        UserID, 
		Caption:       caption,
		Images:        []string{},
		Videos:        []string{},
		LikesCount:    0,
		CommentsCount: 0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Visibility:    visibility,
		Tags:          tags,
	}

	// Upload files to S3
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer file.Close()

		key := "posts/" + fileHeader.Filename
		result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String("sportsify"),
			Key:    aws.String(key),
			Body:   file,
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lowerName := strings.ToLower(fileHeader.Filename)
		if strings.HasSuffix(lowerName, ".jpg") || strings.HasSuffix(lowerName, ".jpeg") || strings.HasSuffix(lowerName, ".png") {
			post.Images = append(post.Images, result.Location)
		} else if strings.HasSuffix(lowerName, ".mp4") || strings.HasSuffix(lowerName, ".mov") {
			post.Videos = append(post.Videos, result.Location)
		}
	}

	// Insert into PostgreSQL
	_, err = db.ExecContext(context.TODO(),
		`INSERT INTO sportsify.posts
		(id, user_id, caption, images, videos, likes_count, comments_count, created_at, updated_at, visibility, tags)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		post.ID,
		post.UserID,
		post.Caption,
		pq.Array(post.Images),
		pq.Array(post.Videos),
		post.LikesCount,
		post.CommentsCount,
		post.CreatedAt,
		post.UpdatedAt,
		post.Visibility,
		pq.Array(post.Tags),
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert failed: " + err.Error()})
		return
	}

	// Return the Post that was stored
	ctx.JSON(http.StatusOK, gin.H{"post": post})
}


func ListUserPostsHandler(ctx *gin.Context, db *sql.DB) {
	// Get user ID from query parameter (or from Auth0 token in real app)
	userID := utils.GetProfileIdFromSession(ctx)
	if userID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
		return
	}

	// Query all posts for this user
	rows, err := db.QueryContext(context.TODO(),
		`SELECT id, user_id, caption, images, videos, likes_count, comments_count, created_at, updated_at, visibility, tags
		 FROM sportsify.posts
		 WHERE user_id = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB query failed: " + err.Error()})
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	for rows.Next() {
		var post models.Post
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Caption,
			pq.Array(&post.Images),
			pq.Array(&post.Videos),
			&post.LikesCount,
			&post.CommentsCount,
			&post.CreatedAt,
			&post.UpdatedAt,
			&post.Visibility,
			pq.Array(&post.Tags),
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB scan failed: " + err.Error()})
			return
		}
		posts = append(posts, post)
	}

	ctx.JSON(http.StatusOK, gin.H{"user_posts": posts})
}
