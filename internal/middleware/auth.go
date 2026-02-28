package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// AuthMiddleware handles cookie-based authentication for the admin dashboard.
type AuthMiddleware struct {
	secret []byte
}

// NewAuthMiddleware creates a new auth middleware with the given secret.
func NewAuthMiddleware(secret string) *AuthMiddleware {
	return &AuthMiddleware{
		secret: []byte(secret),
	}
}

// Require wraps an http.Handler and redirects unauthenticated requests to /login.
func (am *AuthMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("fileorbit_session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		if !am.ValidateToken(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GenerateToken creates an HMAC-based session token.
func (am *AuthMiddleware) GenerateToken() string {
	mac := hmac.New(sha256.New, am.secret)
	mac.Write([]byte("fileorbit_authenticated"))
	return hex.EncodeToString(mac.Sum(nil))
}

// ValidateToken checks if a session token is valid.
func (am *AuthMiddleware) ValidateToken(token string) bool {
	expected := am.GenerateToken()
	return hmac.Equal([]byte(token), []byte(expected))
}

// IsAuthenticated checks if the current request has a valid session cookie.
func (am *AuthMiddleware) IsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("fileorbit_session")
	if err != nil {
		return false
	}
	return am.ValidateToken(cookie.Value)
}
