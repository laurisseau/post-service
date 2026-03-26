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
	"encoding/json"
    _ "github.com/go-sql-driver/mysql"
	"fmt"
)


func scanPost(row *sql.Rows) (models.Post, error) {
    var post models.Post
    var imagesJSON, videosJSON, tagsJSON []byte
    
    err := row.Scan(
        &post.ID,
        &post.UserID,
        &post.Caption,
        &imagesJSON,
        &videosJSON,
        &post.LikesCount,
        &post.CommentsCount,
        &post.CreatedAt,
        &post.UpdatedAt,
        &post.Visibility,
        &tagsJSON,
    )
    if err != nil {
        return post, err
    }
    
    // Unmarshal with null handling
    if len(imagesJSON) > 0 && string(imagesJSON) != "null" {
        if err := json.Unmarshal(imagesJSON, &post.Images); err != nil {
            return post, fmt.Errorf("unmarshal images: %w", err)
        }
    } else {
        post.Images = []string{}
    }
    
    if len(videosJSON) > 0 && string(videosJSON) != "null" {
        if err := json.Unmarshal(videosJSON, &post.Videos); err != nil {
            return post, fmt.Errorf("unmarshal videos: %w", err)
        }
    } else {
        post.Videos = []string{}
    }
    
    if len(tagsJSON) > 0 && string(tagsJSON) != "null" {
        if err := json.Unmarshal(tagsJSON, &post.Tags); err != nil {
            return post, fmt.Errorf("unmarshal tags: %w", err)
        }
    } else {
        post.Tags = []string{}
    }
    
    return post, nil
}

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

	// Get logged-in user ID
	UserID := utils.GetProfileIdFromSession(ctx)

	if UserID == "" {
    ctx.JSON(http.StatusBadRequest, gin.H{"error": "User not logged in"})
    return
	}

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
		}else{
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type: " + fileHeader.Filename})
			return
		}
	}

	// Convert slices to JSON for MySQL
	imagesJSON, _ := json.Marshal(post.Images)
	videosJSON, _ := json.Marshal(post.Videos)
	tagsJSON, _ := json.Marshal(post.Tags)

	/*
	fmt.Println("Post data to insert:")
	fmt.Println("ID:", post.ID)
	fmt.Println("UserID:", post.UserID)
	fmt.Println("Caption:", post.Caption)
	fmt.Println("Images:", string(imagesJSON))   // convert []byte to string
	fmt.Println("Videos:", string(videosJSON))
	fmt.Println("LikesCount:", post.LikesCount)
	fmt.Println("CommentsCount:", post.CommentsCount)
	fmt.Println("CreatedAt:", post.CreatedAt)
	fmt.Println("UpdatedAt:", post.UpdatedAt)
	fmt.Println("Visibility:", post.Visibility)
	fmt.Println("Tags:", string(tagsJSON))
	*/

	// Insert into MySQL
	_, err = db.ExecContext(context.TODO(),
		`INSERT INTO sportsify.posts
		(id, user_id, caption, images, videos, likes_count, comments_count, created_at, updated_at, visibility, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		post.ID,
		post.UserID,
		post.Caption,
		imagesJSON,
		videosJSON,
		post.LikesCount,
		post.CommentsCount,
		post.CreatedAt,
		post.UpdatedAt,
		post.Visibility,
		tagsJSON,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert failed: " + err.Error()})
		return
	}

	// Return the Post that was stored
	ctx.JSON(http.StatusOK, gin.H{"post": post})
}


func ListUserPostsHandler(ctx *gin.Context, db *sql.DB) {
    userID := utils.GetProfileIdFromSession(ctx)
    if userID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id"})
        return
    }

    rows, err := db.QueryContext(context.TODO(),
        `SELECT id, user_id, caption, images, videos, likes_count, comments_count, created_at, updated_at, visibility, tags
         FROM sportsify.posts
         WHERE user_id = ?
         ORDER BY created_at DESC`, userID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB query failed: " + err.Error()})
        return
    }
    defer rows.Close()

    posts := []models.Post{}
    for rows.Next() {
        post, err := scanPost(rows)
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB scan failed: " + err.Error()})
            return
        }
        posts = append(posts, post)
    }

    ctx.JSON(http.StatusOK, gin.H{"user_posts": posts})
}

func DeleteUserPostsHandler(ctx *gin.Context, db *sql.DB) {
    // Get user ID from session
    userID := utils.GetProfileIdFromSession(ctx)
    if userID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User not logged in"})
        return
    }

    // Get post ID from URL parameter
    postID := ctx.Param("id")
    if postID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing post_id"})
        return
    }

    // Use ExecContext for DELETE, not QueryContext
    result, err := db.ExecContext(context.TODO(),
        `DELETE FROM sportsify.posts
        WHERE id = ? AND user_id = ?`, // Use ? for MySQL placeholders
        postID, 
        userID,
    )

    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "DB delete failed: " + err.Error()})
        return
    }

    // Check how many rows were affected
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get rows affected: " + err.Error()})
        return
    }

    if rowsAffected == 0 {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found or you don't have permission to delete it"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "message":       "Post deleted successfully",
        "rows_affected": rowsAffected,
    })
}


func UpdateUserPostHandler(ctx *gin.Context, db *sql.DB) {
    // Get user ID from session
    userID := utils.GetProfileIdFromSession(ctx)
    if userID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "User not logged in"})
        return
    }

    // Get post ID from URL parameter
    postID := ctx.Param("id")
    if postID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing post_id"})
        return
    }

    // Check if request is multipart form (files) or JSON
    contentType := ctx.GetHeader("Content-Type")
    
    if strings.HasPrefix(contentType, "multipart/form-data") {
        // Handle file updates (images/videos)
        updatePostWithFiles(ctx, db, postID, userID)
    } else {
        // Handle JSON updates (text fields only)
        updatePostWithJSON(ctx, db, postID, userID)
    }
}

func updatePostWithJSON(ctx *gin.Context, db *sql.DB, postID, userID string) {
    var updateData struct {
        Caption    *string   `json:"caption"`
        Visibility *string   `json:"visibility"`
        Tags       *[]string `json:"tags"`
    }

    if err := ctx.ShouldBindJSON(&updateData); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
        return
    }

    // Build update query (same as previous version)
    updateFields := []string{}
    updateValues := []interface{}{}

    if updateData.Caption != nil {
        updateFields = append(updateFields, "caption = ?")
        updateValues = append(updateValues, *updateData.Caption)
    }

    if updateData.Visibility != nil {
        validVisibilities := map[string]bool{"public": true, "connections": true, "private": true}
        if !validVisibilities[*updateData.Visibility] {
            ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid visibility"})
            return
        }
        updateFields = append(updateFields, "visibility = ?")
        updateValues = append(updateValues, *updateData.Visibility)
    }

    if updateData.Tags != nil {
        tagsJSON, _ := json.Marshal(*updateData.Tags)
        updateFields = append(updateFields, "tags = ?")
        updateValues = append(updateValues, tagsJSON)
    }

    if len(updateFields) == 0 {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
        return
    }

    updateFields = append(updateFields, "updated_at = ?")
    updateValues = append(updateValues, time.Now())
    updateValues = append(updateValues, postID, userID)

    query := `UPDATE sportsify.posts SET ` + strings.Join(updateFields, ", ") + ` WHERE id = ? AND user_id = ?`
    
    result, err := db.ExecContext(context.TODO(), query, updateValues...)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post: " + err.Error()})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    // Return updated post (fetch from DB)
    updatedPost, err := getPostByID(db, postID, userID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated post"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Post updated successfully", "post": updatedPost})
}

func updatePostWithFiles(ctx *gin.Context, db *sql.DB, postID, userID string) {
    // Parse multipart form
    form, err := ctx.MultipartForm()
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data"})
        return
    }

    files := form.File["files"]
    caption := ""
    if vals, ok := form.Value["caption"]; ok && len(vals) > 0 {
        caption = vals[0]
    }
    
    visibility := "public"
    if vals, ok := form.Value["visibility"]; ok && len(vals) > 0 && vals[0] != "" {
        visibility = vals[0]
    }
    
    tags := []string{}
    if vals, ok := form.Value["tags"]; ok && len(vals) > 0 && vals[0] != "" {
        tags = strings.Split(vals[0], ",")
    }

    // Get existing post to preserve media if no new files are uploaded
    existingPost, err := getPostByID(db, postID, userID)
    if err != nil {
        ctx.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
        return
    }

    // Upload new files if provided
    images := existingPost.Images
    videos := existingPost.Videos

    if len(files) > 0 {
        // Load AWS config and upload files (same as in CreateHandler)
        cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load AWS config"})
            return
        }

        client := s3.NewFromConfig(cfg)
        uploader := manager.NewUploader(client)

        images = []string{}
        videos = []string{}

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
                images = append(images, result.Location)
            } else if strings.HasSuffix(lowerName, ".mp4") || strings.HasSuffix(lowerName, ".mov") {
                videos = append(videos, result.Location)
            } else {
                ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported file type: " + fileHeader.Filename})
                return
            }
        }
    }

    // Convert to JSON
    imagesJSON, _ := json.Marshal(images)
    videosJSON, _ := json.Marshal(videos)
    tagsJSON, _ := json.Marshal(tags)

    // Update post
    _, err = db.ExecContext(context.TODO(),
        `UPDATE sportsify.posts 
         SET caption = ?, images = ?, videos = ?, visibility = ?, tags = ?, updated_at = ?
         WHERE id = ? AND user_id = ?`,
        caption, imagesJSON, videosJSON, visibility, tagsJSON, time.Now(), postID, userID,
    )
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update post: " + err.Error()})
        return
    }

    // Return updated post
    updatedPost, err := getPostByID(db, postID, userID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated post"})
        return
    }

    ctx.JSON(http.StatusOK, gin.H{"message": "Post updated successfully", "post": updatedPost})
}

// Helper function to get a post by ID
func getPostByID(db *sql.DB, postID, userID string) (models.Post, error) {
    var post models.Post
    var imagesJSON, videosJSON, tagsJSON []byte

    err := db.QueryRowContext(context.TODO(),
        `SELECT id, user_id, caption, images, videos, likes_count, comments_count, 
                created_at, updated_at, visibility, tags
         FROM sportsify.posts
         WHERE id = ? AND user_id = ?`,
        postID, userID,
    ).Scan(
        &post.ID,
        &post.UserID,
        &post.Caption,
        &imagesJSON,
        &videosJSON,
        &post.LikesCount,
        &post.CommentsCount,
        &post.CreatedAt,
        &post.UpdatedAt,
        &post.Visibility,
        &tagsJSON,
    )

    if err != nil {
        return post, err
    }

    // Unmarshal JSON fields
    if err := json.Unmarshal(imagesJSON, &post.Images); err != nil {
        post.Images = []string{}
    }
    if err := json.Unmarshal(videosJSON, &post.Videos); err != nil {
        post.Videos = []string{}
    }
    if err := json.Unmarshal(tagsJSON, &post.Tags); err != nil {
        post.Tags = []string{}
    }

    return post, nil
}