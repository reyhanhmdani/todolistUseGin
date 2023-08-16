package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

//func BasicAuth() gin.HandlerFunc {
//	return gin.BasicAuth(gin.Accounts{
//		"key": "value",
//	})
//}

func BasicAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, password, hasAuth := ctx.Request.BasicAuth()
		if !hasAuth || user != "key" || password != "value" {
			//c.Writer.Header().Set("WWW-Authenticate", "Basic realm=Restricted")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  http.StatusUnauthorized,
				"message": "UNAUTHORIZED",
			})
			return
		}
		ctx.Next()
	}
}
