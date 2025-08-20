package middleware

import (
	"net/http"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.

func IsAuthenticated(ctx *gin.Context) {
    session := sessions.Default(ctx)

    if session.Get("profile") == nil {
        ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
            "error": "you have to login",
        })
        return
    }

    ctx.Next() // only runs if authenticated
}
