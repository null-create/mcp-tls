package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrNoAuthHeader   = errors.New("authorization header not provided")
	ErrInvalidToken   = errors.New("invalid token")
	ErrUnauthorized   = errors.New("unauthorized")
	jwtSecret         = []byte("your-secret-key-here") // replace with your own secret!
	ContextUserKeyKey = "user"                         // context key for the parsed claims
)

// Claims is a basic custom claims struct you can extend.
type Claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ParseToken validates the JWT and returns the claims if valid.
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure token method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// Middleware checks the Authorization header and validates the JWT.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, ErrNoAuthHeader.Error(), http.StatusUnauthorized)
			return
		}

		tokenString := extractBearerToken(authHeader)
		if tokenString == "" {
			http.Error(w, ErrInvalidToken.Error(), http.StatusUnauthorized)
			return
		}

		claims, err := ParseToken(tokenString)
		if err != nil {
			http.Error(w, ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		// Pass claims through context
		ctx := context.WithValue(r.Context(), ContextUserKeyKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// extractBearerToken gets the token string from "Authorization: Bearer <token>"
func extractBearerToken(header string) string {
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return ""
}

// CreateToken generates a JWT token with given username and expiry.
func CreateToken(username string, expiry time.Duration) (string, error) {
	claims := &Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// FromContext retrieves claims from context in downstream handlers.
func FromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ContextUserKeyKey).(*Claims)
	return claims, ok
}
