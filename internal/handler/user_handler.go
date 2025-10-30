package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/remiehneppo/material-management/internal/service"
	"github.com/remiehneppo/material-management/types"
)

type UserHandler interface {
	GetUserProfile(ctx *gin.Context)
	UpdateUserProfile(ctx *gin.Context)
	UpdatePassword(ctx *gin.Context)
}

type userHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) UserHandler {
	return &userHandler{
		userService: userService,
	}
}

// getUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the logged-in user
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=types.User} "User profile retrieved successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Failed to retrieve user profile"
// @Security BearerAuth
// @Router /user/profile [get]
func (h *userHandler) GetUserProfile(ctx *gin.Context) {

	user, err := h.userService.GetProfile(ctx)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Failed to retrieve user profile",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "User profile retrieved successfully",
		Data:    user,
	})

}

// updateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile of the logged-in user
// @Tags user
// @Accept json
// @Produce json
// @Param user body types.UpdateUserInfoRequest true "User update info"
// @Success 200 {object} types.Response "User profile updated successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Failed to update user profile"
// @Security BearerAuth
// @Router /user/profile [post]
func (h *userHandler) UpdateUserProfile(ctx *gin.Context) {
	var req types.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request data",
		})
		return
	}

	err := h.userService.Update(ctx, &req)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Failed to update user profile",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "User profile updated successfully",
	})
}

// updatePassword godoc
// @Summary Update user password
// @Description Update the password of the logged-in user
// @Tags user
// @Accept json
// @Produce json
// @Param password body types.UpdateUserPasswordRequest true "User update password info"
// @Success 200 {object} types.Response "User password updated successfully"
// @Failure 400 {object} types.Response "Invalid request data"
// @Failure 500 {object} types.Response "Failed to update user password"
// @Security BearerAuth
// @Router /user/change-password [post]
func (h *userHandler) UpdatePassword(ctx *gin.Context) {
	var req types.UpdateUserPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, types.Response{
			Status:  false,
			Message: "Invalid request data",
		})
		return
	}

	err := h.userService.UpdatePassword(ctx, &req)
	if err != nil {
		ctx.JSON(500, types.Response{
			Status:  false,
			Message: "Failed to update user password",
		})
		return
	}
	ctx.JSON(200, types.Response{
		Status:  true,
		Message: "User password updated successfully",
	})
}
