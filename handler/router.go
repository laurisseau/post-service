package handler

import (
	"database/sql"
	"encoding/gob"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/laurisseau/post-service/handler/post"
	"github.com/laurisseau/user-service/authenticator"
	"github.com/laurisseau/post-service/handler/middleware"
)

// New registers the routes and returns the router.
func Router(db *sql.DB, auth *authenticator.Authenticator, router *gin.Engine) {
	
	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	// Create a /posts route group
    postsGroup := router.Group("/posts")
    {
        postsGroup.POST("/create", middleware.MetricsMiddleware(), middleware.IsAuthenticated, func(c *gin.Context) {
            post.CreateHandler(c, db)
        })

        postsGroup.GET("/listUserPosts", middleware.MetricsMiddleware(), middleware.IsAuthenticated, func(c *gin.Context) {
            post.ListUserPostsHandler(c, db)
        })

        postsGroup.DELETE("/delete/:id", middleware.MetricsMiddleware(), middleware.IsAuthenticated, func(c *gin.Context) {
            post.DeleteUserPostsHandler(c, db)
        })

        postsGroup.PATCH("/update/:id", middleware.MetricsMiddleware(), middleware.IsAuthenticated, func(c *gin.Context) {
            post.UpdateUserPostHandler(c, db)
        })

    }
}
