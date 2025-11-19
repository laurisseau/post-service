package main

import (
    "log"
	"net/http"
	"github.com/laurisseau/post-service/handler"
    "github.com/laurisseau/user-service/authenticator"
	"github.com/gin-gonic/gin"
)

func main() {

    r := gin.Default()

    // Initialize Authenticator
	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize Authenticator: %v", err)
	}

	handler.Router(auth, r)

    r.GET("post/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Welcome to Sportsify post!",
        })
    })

    r.Run(":8081") // Starts server on http://localhost:8081 post application port
}