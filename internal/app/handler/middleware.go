package handler

import (
	"net/http"

	"WEB/internal/app/role"

	"github.com/gin-gonic/gin"
)

// RoleMiddleware middleware для проверки ролей
func (h *Handler) RoleMiddleware(allowedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userRole, exists := ctx.Get("user_role")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User role not found"})
			ctx.Abort()
			return
		}

		hasAccess := false
		for _, allowedRole := range allowedRoles {
			if userRole.(role.Role) == allowedRole {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
