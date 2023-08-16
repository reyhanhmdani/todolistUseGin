package cfg

import (
	"github.com/dgrijalva/jwt-go"
	"os"
	"strconv"
	"time"
)

var JwtKey = []byte(os.Getenv("JWT_PRIVATE_KEY"))

// Simpan token yang sudah digunakan dalam map
var UsedTokens = make(map[string]bool)

// payload untuk token
type Claims struct {
	Username string `json:"username"`
	UserID   int64  `json:"user_id"`
	jwt.StandardClaims
}

// fungsi untuk membuat token
func CreateToken(username string, userID int64) (string, error) {
	tokenTTL, _ := strconv.Atoi(os.Getenv("TOKEN_TTL"))
	// mengatur waktu kadaluwarsa token
	//tokenTTLMinutes := tokenTTL / 60
	expirationTime := time.Now().Add(time.Minute * time.Duration(tokenTTL))

	// membuat claims
	claims := &Claims{
		Username: username,
		UserID:   userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	// membuat token dengan signing method HS256 dan secret key
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)

	if err != nil {
		return "", err
	}

	// simpan token yang di generate
	//UsedTokens[tokenString] = false

	return tokenString, nil
}

//func ParseToken(tokenString string) (string, int, error) {
//	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//		// Validasi tipe algoritma dan return secret key
//		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
//		}
//		return JwtKey, nil
//	})
//
//	if err != nil {
//		return "", 0, err
//	}
//
//	claims, ok := token.Claims.(jwt.MapClaims)
//	if !ok || !token.Valid {
//		return "", 0, fmt.Errorf("Invalid token")
//	}
//
//	username, ok := claims["username"].(string)
//	if !ok {
//		return "", 0, fmt.Errorf("Invalid token")
//	}
//
//	userID, ok := claims["user_id"].(float64)
//	if !ok {
//		return "", 0, fmt.Errorf("Invalid token")
//	}
//
//	return username, int(userID), nil
//}
