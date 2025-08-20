package post

import (
	"net/http"
	"github.com/gin-gonic/gin"
	
	/*
	"encoding/json"
	"log"
	"github.com/laurisseau/sportsify-config"
	"github.com/laurisseau/user-service/utils"
	"github.com/laurisseau/user-service/models"
	"github.com/laurisseau/user-service/authenticator"
	*/
)




// Handler to get profile information from the session.
func createHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"post": "Post created",
	})
}
