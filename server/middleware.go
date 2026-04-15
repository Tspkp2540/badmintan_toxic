package main

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userId"
const userRoleKey contextKey = "userRole"

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "badminton-hub-secret-key-change-in-production"
	}
	return []byte(secret)
}

type jwtClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func signToken(userID, email string) (string, error) {
	claims := jwtClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "ไม่ได้เข้าสู่ระบบ")
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return getJWTSecret(), nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "Token ไม่ถูกต้องหรือหมดอายุ")
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)

		// Fetch user role from DB and add to context
		var role string
		err = sqlDB.QueryRow("SELECT role FROM users WHERE id = ?", claims.UserID).Scan(&role)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "ไม่พบผู้ใช้")
			return
		}
		ctx = context.WithValue(ctx, userRoleKey, role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// optionalAuthMiddleware extracts user info if token present, but doesn't block
func optionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			claims := &jwtClaims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return getJWTSecret(), nil
			})
			if err == nil && token.Valid {
				ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
				var role string
				if sqlDB.QueryRow("SELECT role FROM users WHERE id = ?", claims.UserID).Scan(&role) == nil {
					ctx = context.WithValue(ctx, userRoleKey, role)
				}
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// requireRole creates middleware that requires one of the specified roles
func requireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := getUserRole(r)
			for _, allowed := range roles {
				if role == allowed {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeError(w, http.StatusForbidden, "คุณไม่มีสิทธิ์เข้าถึง")
		})
	}
}

func getUserID(r *http.Request) string {
	if id, ok := r.Context().Value(userIDKey).(string); ok {
		return id
	}
	return ""
}

func getUserRole(r *http.Request) string {
	if role, ok := r.Context().Value(userRoleKey).(string); ok {
		return role
	}
	return ""
}
