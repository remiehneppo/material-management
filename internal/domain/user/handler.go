package user

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	domainsession "github.com/remiehneppo/material-management/internal/domain/session"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Handler struct {
	client     *mongo.Client
	collection *mongo.Collection
	sessions   *mongo.Collection
}

func NewHandler(client *mongo.Client, db *mongo.Database) *Handler {
	return &Handler{client: client, collection: db.Collection("users"), sessions: db.Collection("user_sessions")}
}

func current(ctx *gin.Context) (*types.User, bson.ObjectID, bool) {
	claims, ok := ctx.Value("user").(*types.User)
	if !ok {
		return nil, bson.NilObjectID, false
	}
	id, err := bson.ObjectIDFromHex(claims.ID)
	return claims, id, err == nil
}

// GetProfile godoc
// @Summary Get current user profile
// @Tags user
// @Produce json
// @Success 200 {object} types.Response{data=types.User}
// @Security BearerAuth
// @Router /user/profile [get]
func (h *Handler) GetProfile(ctx *gin.Context) {
	_, id, ok := current(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrUnauthorized.Error()})
		return
	}
	var user types.User
	if err := h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		ctx.JSON(http.StatusNotFound, types.Response{Status: false, Message: "Không tìm thấy thông tin người dùng."})
		return
	}
	user.Password = ""
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã tải thông tin người dùng.", Data: user})
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Tags user
// @Accept json
// @Produce json
// @Param request body types.UpdateUserInfoRequest true "Profile changes"
// @Success 200 {object} types.Response
// @Security BearerAuth
// @Router /user/profile [post]
func (h *Handler) UpdateProfile(ctx *gin.Context) {
	_, id, ok := current(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrUnauthorized.Error()})
		return
	}
	var request types.UpdateUserInfoRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Thông tin người dùng không hợp lệ."})
		return
	}
	set := bson.M{}
	if request.FullName != "" {
		set["full_name"] = request.FullName
	}
	if request.Workspace != "" {
		set["workspace"] = request.Workspace
	}
	if request.WorkspaceRole != "" {
		set["workspace_role"] = request.WorkspaceRole
	}
	if len(set) > 0 {
		if _, err := h.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": set}); err != nil {
			ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể cập nhật thông tin người dùng."})
			return
		}
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã cập nhật thông tin người dùng."})
}

// ChangePassword godoc
// @Summary Change current user password
// @Tags user
// @Accept json
// @Produce json
// @Param request body types.UpdateUserPasswordRequest true "Password change"
// @Success 200 {object} types.Response
// @Security BearerAuth
// @Router /user/change-password [post]
func (h *Handler) ChangePassword(ctx *gin.Context) {
	_, id, ok := current(ctx)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrUnauthorized.Error()})
		return
	}
	var request types.UpdateUserPasswordRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: "Thông tin đổi mật khẩu không hợp lệ."})
		return
	}
	if !regexp.MustCompile(types.PASSWORD_REGEX).MatchString(request.NewPassword) {
		ctx.JSON(http.StatusBadRequest, types.Response{Status: false, Message: types.ErrPasswordInvalid.Error()})
		return
	}
	var user types.User
	if err := h.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user); err != nil {
		ctx.JSON(http.StatusNotFound, types.Response{Status: false, Message: "Không tìm thấy thông tin người dùng."})
		return
	}
	valid := user.Password == request.OldPassword
	if len(user.Password) >= 10 && user.Password[:10] == "$argon2id$" {
		valid, _ = domainsession.VerifyPassword(user.Password, request.OldPassword)
	}
	if !valid {
		ctx.JSON(http.StatusUnauthorized, types.Response{Status: false, Message: types.ErrPasswordIncorrect.Error()})
		return
	}
	hash, err := domainsession.HashPassword(request.NewPassword)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể mã hóa mật khẩu mới."})
		return
	}
	session, err := h.client.StartSession()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể đổi mật khẩu."})
		return
	}
	defer session.EndSession(ctx)
	_, err = session.WithTransaction(ctx, func(tx context.Context) (interface{}, error) {
		if _, err := h.collection.UpdateOne(tx, bson.M{"_id": id}, bson.M{"$set": bson.M{"password": hash}}); err != nil {
			return nil, err
		}
		if _, err := h.sessions.UpdateMany(tx, bson.M{"user_id": user.ID, "revoked_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{"revoked_at": time.Now().Unix()}}); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, types.Response{Status: false, Message: "Không thể đổi mật khẩu."})
		return
	}
	ctx.JSON(http.StatusOK, types.Response{Status: true, Message: "Đã đổi mật khẩu."})
}
