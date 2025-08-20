package post

import (
	"context"
	"net/http"
	"strings"
	"github.com/laurisseau/post-service/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	/*
		"encoding/json"
		"log"
		"github.com/laurisseau/sportsify-config"
		"github.com/laurisseau/user-service/utils"
		"github.com/laurisseau/user-service/models"
		"github.com/laurisseau/user-service/authenticator"
	*/)

func CreateHandler(ctx *gin.Context) {
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

    // Load AWS config (automatically reads from env vars, ~/.aws, etc.)
    cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load AWS config"})
        return
    }

    client := s3.NewFromConfig(cfg)
    uploader := manager.NewUploader(client)

    //var post Post

    post := models.Post{}

    for _, fileHeader := range files {
        file, err := fileHeader.Open()
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
            return
        }
        defer file.Close()

        key := "posts/" + fileHeader.Filename

        // Upload with context
        result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
            Bucket: aws.String("sportsify"),
            Key:    aws.String(key),
            Body:   file,
        })
        if err != nil {
            ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        // Classify into images or videos
        if strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".jpg") ||
            strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".jpeg") ||
            strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".png") {
            post.Images = append(post.Images, result.Location) // result.Location is in v2 uploader
        } else if strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".mp4") ||
            strings.HasSuffix(strings.ToLower(fileHeader.Filename), ".mov") {
            post.Videos = append(post.Videos, result.Location)
        }
    }

    ctx.JSON(http.StatusOK, gin.H{"post": post})
}

