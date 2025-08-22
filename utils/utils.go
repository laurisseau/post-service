package utils

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// GetProfileFromSession retrieves the user profile from the session.
func GetProfileFromSession(ctx *gin.Context) interface{} {
	session := sessions.Default(ctx)
	return session.Get("profile")
}

func GetProfileIdFromSession(ctx *gin.Context) string {
	profile := GetProfileFromSession(ctx)
	if profile == nil {
		return ""
	}

	if profileMap, ok := profile.(map[string]interface{}); ok {
		if id, exists := profileMap["sub"]; exists {
			return id.(string)
		}
	}

	return ""
}