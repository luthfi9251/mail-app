package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"

	"email-app/config"
	"email-app/database"
	"email-app/models"
)

const (
	CookieName = "token"
	TokenTTL   = 7 * 24 * time.Hour
)

type Claims struct {
	UserID       string `json:"userId"`
	TokenVersion int    `json:"tv"`
	jwt.RegisteredClaims
}

// IssueToken signs a session token for the user and sets it as the auth cookie.
func IssueToken(c *gin.Context, user models.User) error {
	claims := Claims{
		UserID:       user.ID.Hex(),
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenTTL)),
		},
	}
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(config.JWTSecret)
	if err != nil {
		return err
	}
	setAuthCookie(c, tokenString, int(TokenTTL.Seconds()))
	return nil
}

func ClearToken(c *gin.Context) {
	setAuthCookie(c, "", -1)
}

func setAuthCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(CookieName, value, maxAge, "/", "", config.CookieSecure, true)
}

// AuthRequired validates the session cookie and re-reads the user from the
// database on every request, so deleted users, role changes and password
// changes (which bump tokenVersion) take effect immediately.
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(CookieName)
		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return config.JWTSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		userID, err := bson.ObjectIDFromHex(claims.UserID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		var user models.User
		if err := database.GetCollection("users").FindOne(c.Request.Context(), bson.M{"_id": userID}).Decode(&user); err != nil ||
			user.TokenVersion != claims.TokenVersion {
			ClearToken(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired"})
			return
		}

		c.Set("userId", user.ID.Hex())
		c.Set("email", user.Email)
		c.Set("role", user.Role)
		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
