package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	domainsession "github.com/remiehneppo/material-management/internal/domain/session"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

const BearerPrefix = "bearer "

type AuthMiddleware struct {
	jwtService *service.JWTService
	sessions   *domainsession.Manager
}

func NewAuthMiddleware(jwtService *service.JWTService, sessions *domainsession.Manager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		sessions:   sessions,
	}
}

func (a *AuthMiddleware) AuthBearerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		accessToken := ctx.GetHeader("Authorization")
		if accessToken == "" {
			res := types.Response{
				Status:  false,
				Message: "Yêu cầu chưa có thông tin xác thực.",
			}
			ctx.JSON(http.StatusUnauthorized, res)
			ctx.Abort()
			return
		}

		// Check if the token has the Bearer prefix
		if len(accessToken) < len(BearerPrefix) || strings.ToLower(accessToken[:len(BearerPrefix)]) != BearerPrefix {
			res := types.Response{
				Status:  false,
				Message: "Thông tin xác thực không đúng định dạng.",
			}
			ctx.JSON(http.StatusUnauthorized, res)
			ctx.Abort()
			return
		}

		// Remove "Bearer " prefix
		accessToken = accessToken[len(BearerPrefix):]

		user, err := a.jwtService.ValidateAccessToken(accessToken)
		if err != nil {
			res := types.Response{
				Status:  false,
				Message: "Phiên đăng nhập không hợp lệ.",
			}
			ctx.JSON(http.StatusUnauthorized, res)
			ctx.Abort()
			return
		}
		if err := a.sessions.ValidateAccessSession(ctx, user.SessionID, user.ID); err != nil {
			ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: "Phiên đăng nhập đã hết hiệu lực."})
			ctx.Abort()
			return
		}

		ctx.Set("user", user)
		ctx.Next()
	}
}
