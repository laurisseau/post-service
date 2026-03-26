package main

import (
    "log"
	"net/http"
	"github.com/laurisseau/post-service/handler"
    "github.com/laurisseau/user-service/authenticator"
	"github.com/gin-gonic/gin"
    "github.com/laurisseau/sportsify-config"
    //"github.com/laurisseau/post-service/metrics"
    "github.com/laurisseau/post-service/handler/middleware"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

    r := gin.Default()

    db, err := config.DB()
    if err != nil {
        panic(err)
    }

    // Initialize Authenticator
	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize Authenticator: %v", err)
	}

	handler.Router(db, auth, r)

    // Add metrics middleware
    r.Use(middleware.MetricsMiddleware())

    r.GET("posts/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Welcome to Sportsify post!",
        })
    })

    // Prometheus metrics endpoint
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))

    r.Run(":8081") // Starts server on http://localhost:8081/posts post application port
}