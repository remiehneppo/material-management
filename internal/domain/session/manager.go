package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/remiehneppo/material-management/types"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const sessionsCollection = "user_sessions"

type accessTokenIssuer interface {
	GenerateAccessTokenForSession(user *types.User, sessionID string) (string, error)
}

type document struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    string        `bson:"user_id"`
	TokenHash string        `bson:"token_hash"`
	CreatedAt int64         `bson:"created_at"`
	ExpiresAt int64         `bson:"expires_at"`
	RevokedAt int64         `bson:"revoked_at,omitempty"`
}

type Tokens struct {
	AccessToken  string
	RefreshToken string
}

type Manager struct {
	db         *mongo.Database
	access     accessTokenIssuer
	refreshTTL time.Duration
	now        func() time.Time
}

func NewManager(db *mongo.Database, access accessTokenIssuer, refreshTTL time.Duration) *Manager {
	if refreshTTL <= 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &Manager{db: db, access: access, refreshTTL: refreshTTL, now: time.Now}
}

func (m *Manager) Login(ctx context.Context, username, password string) (*Tokens, error) {
	if !regexp.MustCompile(types.USERNAME_REGEX).MatchString(username) || !regexp.MustCompile(types.PASSWORD_REGEX).MatchString(password) {
		return nil, types.ErrUsernameOrPasswordNotCorrect
	}
	var user types.User
	if err := m.db.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return nil, types.ErrUsernameOrPasswordNotCorrect
	}
	valid := false
	if strings.HasPrefix(user.Password, "$argon2id$") {
		valid, _ = VerifyPassword(user.Password, password)
	} else {
		valid = subtle.ConstantTimeCompare([]byte(user.Password), []byte(password)) == 1
		if valid {
			hash, err := HashPassword(password)
			if err != nil {
				return nil, err
			}
			if _, err := m.db.Collection("users").UpdateOne(ctx, bson.M{"_id": mustObjectID(user.ID)}, bson.M{"$set": bson.M{"password": hash}}); err != nil {
				return nil, err
			}
		}
	}
	if !valid {
		return nil, types.ErrUsernameOrPasswordNotCorrect
	}
	return m.create(ctx, &user)
}

func (m *Manager) create(ctx context.Context, user *types.User) (*Tokens, error) {
	id := bson.NewObjectID()
	secret, hash, err := newSecret()
	if err != nil {
		return nil, err
	}
	now := m.now()
	doc := document{ID: id, UserID: user.ID, TokenHash: hash, CreatedAt: now.Unix(), ExpiresAt: now.Add(m.refreshTTL).Unix()}
	if _, err := m.db.Collection(sessionsCollection).InsertOne(ctx, doc); err != nil {
		return nil, err
	}
	access, err := m.access.GenerateAccessTokenForSession(user, id.Hex())
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: access, RefreshToken: id.Hex() + "." + secret}, nil
}

func (m *Manager) Refresh(ctx context.Context, token string) (*Tokens, error) {
	id, secret, err := parseToken(token)
	if err != nil {
		return nil, types.ErrUnauthorized
	}
	newSecretValue, newHash, err := newSecret()
	if err != nil {
		return nil, err
	}
	var session document
	err = m.db.Collection(sessionsCollection).FindOneAndUpdate(ctx, bson.M{
		"_id": id, "token_hash": hashSecret(secret), "revoked_at": bson.M{"$exists": false}, "expires_at": bson.M{"$gt": m.now().Unix()},
	}, bson.M{"$set": bson.M{"token_hash": newHash}}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&session)
	if err != nil {
		return nil, types.ErrUnauthorized
	}
	var user types.User
	if err := m.db.Collection("users").FindOne(ctx, bson.M{"_id": mustObjectID(session.UserID)}).Decode(&user); err != nil {
		return nil, types.ErrUnauthorized
	}
	access, err := m.access.GenerateAccessTokenForSession(&user, id.Hex())
	if err != nil {
		return nil, err
	}
	return &Tokens{AccessToken: access, RefreshToken: id.Hex() + "." + newSecretValue}, nil
}

func (m *Manager) Logout(ctx context.Context, token string) error {
	id, secret, err := parseToken(token)
	if err != nil {
		return nil
	}
	_, err = m.db.Collection(sessionsCollection).UpdateOne(ctx, bson.M{"_id": id, "token_hash": hashSecret(secret)}, bson.M{"$set": bson.M{"revoked_at": m.now().Unix()}})
	return err
}

func (m *Manager) ValidateAccessSession(ctx context.Context, sessionID, userID string) error {
	id, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return types.ErrUnauthorized
	}
	err = m.db.Collection(sessionsCollection).FindOne(ctx, bson.M{
		"_id": id, "user_id": userID, "revoked_at": bson.M{"$exists": false}, "expires_at": bson.M{"$gt": m.now().Unix()},
	}).Err()
	if err != nil {
		return types.ErrUnauthorized
	}
	return nil
}

func newSecret() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}
	secret := base64.RawURLEncoding.EncodeToString(raw)
	return secret, hashSecret(secret), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func parseToken(token string) (bson.ObjectID, string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 || parts[1] == "" {
		return bson.NilObjectID, "", errors.New("invalid refresh token")
	}
	id, err := bson.ObjectIDFromHex(parts[0])
	return id, parts[1], err
}

func mustObjectID(id string) bson.ObjectID {
	objectID, _ := bson.ObjectIDFromHex(id)
	return objectID
}
