package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/laurisseau/user-service/authenticator"
    "encoding/gob"
    "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
    "github.com/laurisseau/post-service/handler/post"
	"github.com/laurisseau/sportsify-config"
    //"github.com/laurisseau/post-service/handler/middleware"
    
)

// New registers the routes and returns the router.
func Router(auth *authenticator.Authenticator, router *gin.Engine) {
	
	db := config.DB()
    println("db result", db)
	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.POST("/post/create", func(c *gin.Context) {
		post.CreateHandler(c, db)
	})

	router.GET("/post/listUserPosts", func(c *gin.Context) {
		post.ListUserPostsHandler(c, db)
	})

	/*
	router.DELETE("/post/delete:id", middleware.IsAuthenticated,)
    router.PATCH("/post/update:id", middleware.IsAuthenticated,)
    
	*/
}
