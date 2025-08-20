package handler

import (
    "github.com/gin-gonic/gin"
    "github.com/laurisseau/user-service/authenticator"
    "encoding/gob"
    "database/sql"
    "github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
    "github.com/laurisseau/post-service/handler/post"
    "github.com/laurisseau/post-service/handler/middleware"
    
)

// New registers the routes and returns the router.
func Router(db *sql.DB, auth *authenticator.Authenticator, router *gin.Engine) {

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.POST("/post/create", middleware.IsAuthenticated, post.CreateHandler)

	/*
	router.DELETE("/post/delete:id", middleware.IsAuthenticated,)
    router.PATCH("/post/update:id", middleware.IsAuthenticated,)
    router.GET("/post/get:id", middleware.IsAuthenticated, )
	*/
}
