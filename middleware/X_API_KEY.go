package middleware

import "github.com/gin-gonic/gin"

func XAPIKEY() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.GetHeader("Authorization")
		if key != "secret_Key" {
			ctx.AbortWithStatusJSON(401, gin.H{
				"message": "UNAUTHORIZED",
			})
			return
		}
	}
}
