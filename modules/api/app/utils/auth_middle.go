package utils

import (
	"github.com/signmem/falcon-plus/modules/api/config"
	"net/http"

	"github.com/gin-gonic/gin"
	h "github.com/signmem/falcon-plus/modules/api/app/helper"
	"github.com/spf13/viper"
)

func AuthSessionMidd(c *gin.Context) {
	auth, err := h.SessionChecking(c)
	if !viper.GetBool("skip_auth") {
		if err != nil || auth != true {
			config.Logger.Errorf("error: %v, auth: %v", err.Error(), auth)
			c.Set("auth", auth)
			h.JSONR(c, http.StatusUnauthorized, err)
			c.Abort()
			return
		}
	}
	c.Set("auth", auth)
}

func CORS() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		context.Writer.Header().Set("Access-Control-Max-Age", "86400")
		context.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		context.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Apitoken")
		context.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		context.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if context.Request.Method == "OPTIONS" {
			context.AbortWithStatus(200)
		} else {
			context.Next()
		}
	}
}
