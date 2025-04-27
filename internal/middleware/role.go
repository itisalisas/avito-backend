package middleware

import (
	"github.com/itisalisas/avito-backend/internal/generated/dto"
	"net/http"
)

func CheckRole(requiredRoles ...dto.PostRegisterJSONBodyRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := dto.PostRegisterJSONBodyRole(r.Context().Value(userRoleKey).(string))

			for _, requiredRole := range requiredRoles {
				if role == requiredRole {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
