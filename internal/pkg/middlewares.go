package pkg

import (
	"errors"
	"net/http"
	"strings"

	"WEB/internal/app/ds"
	"WEB/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/sirupsen/logrus"
)

const jwtPrefix = "Bearer "

// WithAuthCheck middleware с проверкой ролей
func (a *Application) WithAuthCheck(assignedRoles ...role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jwtStr := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(jwtStr, jwtPrefix) {
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Отрезаем префикс
		jwtStr = jwtStr[len(jwtPrefix):]

		// Проверяем JWT в блеклисте Redis
		err := a.RedisClient.CheckJWTInBlacklist(ctx.Request.Context(), jwtStr)
		if err == nil {
			// Токен в блеклисте
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		if !errors.Is(err, redis.Nil) {
			// Внутренняя ошибка Redis
			logrus.Errorf("Redis error: %v", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		// Парсим JWT токен
		token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.Config.JWT.Secret), nil
		})
		if err != nil {
			logrus.WithError(err).Warn("JWT parsing failed")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !token.Valid {
			logrus.Warn("Invalid JWT token")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		myClaims, ok := token.Claims.(*ds.JWTClaims)
		if !ok {
			logrus.Warn("Invalid JWT claims")
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Проверяем роли, если они указаны
		if len(assignedRoles) > 0 {
			hasAccess := false
			for _, assignedRole := range assignedRoles {
				if myClaims.Role == assignedRole {
					hasAccess = true
					break
				}
			}

			if !hasAccess {
				logrus.Warnf("Role %s is not assigned in %v", myClaims.Role, assignedRoles)
				ctx.AbortWithStatus(http.StatusForbidden)
				return
			}
		}

		// Сохраняем claims в контекст
		ctx.Set("jwt_claims", myClaims)
		ctx.Set("user_uuid", myClaims.UserUUID)
		ctx.Set("user_role", myClaims.Role)

		logrus.Debugf("User %s with role %s authenticated", myClaims.UserUUID, myClaims.Role)
		ctx.Next()
	}
}
