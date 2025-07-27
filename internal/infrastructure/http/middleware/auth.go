package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/radiophysiker/d56/internal/domain/user"
	"github.com/radiophysiker/d56/internal/infrastructure/jwt"
)

// contextKey используется для типобезопасного хранения значений в контексте
type contextKey string

const UserIDKey contextKey = "user_id"

// AuthMiddleware проверяет JWT токен и добавляет UserID в контекст
type AuthMiddleware struct {
	jwtService *jwt.Service
}

// NewAuthMiddleware создает новый middleware для аутентификации
func NewAuthMiddleware(jwtService *jwt.Service) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth проверяет JWT токен и добавляет UserID в контекст
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header is required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.jwtService.ParseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Добавляем UserID в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext извлекает UserID из контекста
// Эта функция находится в Infrastructure слое, так как знает о HTTP контексте
func GetUserIDFromContext(ctx context.Context) (user.UserID, bool) {
	userID, ok := ctx.Value(UserIDKey).(user.UserID)
	return userID, ok
}
