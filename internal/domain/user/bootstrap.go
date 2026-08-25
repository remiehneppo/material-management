package user

import (
	"context"
	"fmt"
	"regexp"
	"time"

	domainsession "github.com/remiehneppo/material-management/internal/domain/session"
	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type BootstrapAdminConfig struct {
	Username  string
	Password  string
	FullName  string
	Workspace string
}

// BootstrapAdmin creates the initial administrator when bootstrap credentials
// are configured. Existing users are never modified.
func BootstrapAdmin(ctx context.Context, db *mongo.Database, cfg BootstrapAdminConfig) (bool, error) {
	if cfg.Username == "" && cfg.Password == "" {
		return false, nil
	}
	if cfg.Username == "" || cfg.Password == "" {
		return false, fmt.Errorf("both bootstrap admin username and password must be configured")
	}
	if !regexp.MustCompile(types.USERNAME_REGEX).MatchString(cfg.Username) {
		return false, fmt.Errorf("bootstrap admin username is invalid")
	}
	if !regexp.MustCompile(types.PASSWORD_REGEX).MatchString(cfg.Password) {
		return false, fmt.Errorf("bootstrap admin password is invalid")
	}
	if db == nil {
		return false, fmt.Errorf("database is required to bootstrap admin")
	}

	users := db.Collection("users")
	if _, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetName("unique_username").SetUnique(true).
			SetPartialFilterExpression(bson.M{"username": bson.M{"$type": "string"}}),
	}); err != nil {
		return false, fmt.Errorf("ensure unique username index: %w", err)
	}

	hash, err := domainsession.HashPassword(cfg.Password)
	if err != nil {
		return false, fmt.Errorf("hash bootstrap admin password: %w", err)
	}
	now := time.Now().Unix()
	result, err := users.UpdateOne(ctx,
		bson.M{"username": cfg.Username},
		bson.M{"$setOnInsert": types.User{
			Username:      cfg.Username,
			Password:      hash,
			FullName:      cfg.FullName,
			WorkspaceRole: types.USER_ROLE_ADMIN,
			Workspace:     cfg.Workspace,
			CreateAt:      now,
			UpdateAt:      now,
		}},
		options.UpdateOne().SetUpsert(true),
	)
	if err != nil {
		return false, fmt.Errorf("create bootstrap admin: %w", err)
	}
	return result.UpsertedCount == 1, nil
}
