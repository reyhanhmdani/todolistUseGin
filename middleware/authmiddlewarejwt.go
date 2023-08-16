package middleware

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"todoGin/cfg"
	"todoGin/model/respErr"
)

// secret key untuk signing token
// middleware konsep nya adalah sesuatu yang ibaratnya intercept , request -> server,
func Authmiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// mengambil token dari header Authorization
		authHeader := ctx.GetHeader("Authorization")

		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &respErr.ErrorResponse{
				Message: "Unauthorized",
				Status:  http.StatusUnauthorized,
			})
			return
		}

		// split token dari header
		tokenString := authHeader[len("Bearer "):]

		// parsing token dengan secret key
		claims := &cfg.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return cfg.JwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, &respErr.ErrorResponse{
					Message: "Unauthorized",
					Status:  http.StatusUnauthorized,
				})
				return
			}
			ctx.AbortWithStatusJSON(http.StatusBadRequest, &respErr.ErrorResponse{
				Message: "invalid or expired token",
				Status:  http.StatusBadRequest,
			})
			return
		}
		// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6IlJleSIsInVzZXJfaWQiOjEsImV4cCI6MTY5MTY3MDAwOH0.vcOO1AXA1gonaU_EChbAC1fakp3Gr5UcrZCAZofrhws

		if !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, &respErr.ErrorResponse{
				Message: "Unauthorized (non Valid)",
				Status:  http.StatusUnauthorized,
			})
			return
		}

		//// Cek apakah token sudah digunakan sebelumnya
		//if cfg.UsedTokens[tokenString] {
		//	ctx.AbortWithStatusJSON(http.StatusUnauthorized, &respErr.ErrorResponse{
		//		Message: "Token has already been used",
		//		Status:  http.StatusUnauthorized,
		//	})
		//	return
		//}
		//// Tandai token sebagai digunakan
		//cfg.UsedTokens[tokenString] = true

		// token valid, ambil username dari claim dan simpan ke dalam konteks
		ctx.Set("username", claims.Username)
		ctx.Set("user_id", claims.UserID)

		// token valid, melanjutkan ke handler
		ctx.Next()

	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				logrus.Error("Panic occurred:", r)
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, respErr.ErrorResponse{
					Message: "Internal Server Error",
					Status:  http.StatusInternalServerError,
				})
			}
		}()

		ctx.Next()
	}
}
