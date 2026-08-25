package session

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/types"
)

const RefreshCookieName = "refresh_token"

type Handler struct {
	manager    *Manager
	production bool
	loginGuard *loginGuard
}

func NewHandler(manager *Manager, production bool) *Handler {
	return &Handler{manager: manager, production: production, loginGuard: newLoginGuard()}
}

// Login godoc
// @Summary Login and create a User Session
// @Tags auth
// @Accept json
// @Produce json
// @Param request body types.LoginRequest true "Credentials"
// @Success 200 {object} types.Response{data=types.LoginResponse}
// @Router /auth/login [post]
func (h *Handler) Login(ctx *gin.Context) {
	var request types.LoginRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Vui lòng nhập tên đăng nhập và mật khẩu."})
		return
	}
	if !h.loginGuard.allow(ctx.ClientIP(), request.Username) || !h.loginGuard.acquire() {
		ctx.JSON(http.StatusTooManyRequests, types.Response{Status: false, Message: "Bạn đã đăng nhập sai quá nhiều lần. Vui lòng thử lại sau."})
		return
	}
	defer h.loginGuard.release()
	tokens, err := h.manager.Login(ctx, request.Username, request.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: err.Error()})
		return
	}
	h.setRefreshCookie(ctx, tokens.RefreshToken)
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đăng nhập thành công.", Data: types.LoginResponse{AccessToken: tokens.AccessToken}})
}

// Refresh godoc
// @Summary Rotate the refresh cookie and return a new access token
// @Tags auth
// @Produce json
// @Success 200 {object} types.Response{data=types.LoginResponse}
// @Router /auth/refresh [post]
func (h *Handler) Refresh(ctx *gin.Context) {
	token, err := ctx.Cookie(RefreshCookieName)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrUnauthorized.Error()})
		return
	}
	tokens, err := h.manager.Refresh(ctx, token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: err.Error()})
		return
	}
	h.setRefreshCookie(ctx, tokens.RefreshToken)
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã gia hạn phiên đăng nhập.", Data: types.LoginResponse{AccessToken: tokens.AccessToken}})
}

// Logout godoc
// @Summary Revoke the current User Session
// @Tags auth
// @Produce json
// @Success 200 {object} types.Response
// @Router /auth/logout [post]
func (h *Handler) Logout(ctx *gin.Context) {
	if token, err := ctx.Cookie(RefreshCookieName); err == nil {
		if err := h.manager.Logout(ctx, token); err != nil {
			h.clearRefreshCookie(ctx)
			ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể kết thúc phiên đăng nhập."})
			return
		}
	}
	h.clearRefreshCookie(ctx)
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã đăng xuất."})
}

func (h *Handler) setRefreshCookie(ctx *gin.Context, value string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(RefreshCookieName, value, int((7 * 24 * 60 * 60)), "/api/v1/auth", "", h.secure(ctx), true)
}

func (h *Handler) clearRefreshCookie(ctx *gin.Context) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(RefreshCookieName, "", -1, "/api/v1/auth", "", h.secure(ctx), true)
}

func (h *Handler) secure(ctx *gin.Context) bool {
	host := strings.Split(ctx.Request.Host, ":")[0]
	return h.production || (host != "localhost" && host != "127.0.0.1" && host != "::1")
}
