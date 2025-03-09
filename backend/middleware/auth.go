package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
)

// AuthMiddleware Middleware to validate JWT
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}
		var jwtSecret = []byte("4b08286222a69603f47f766d7f95a93d") // Replace with a secure secret

		// Validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Attach user data to request context
		claims := token.Claims.(jwt.MapClaims)
		r = r.WithContext(context.WithValue(r.Context(), "user", claims))

		next.ServeHTTP(w, r)
	})
}
