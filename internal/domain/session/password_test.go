package session

import (
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestArgon2idPasswordRoundTrip(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("unexpected hash: %s", hash)
	}
	valid, err := VerifyPassword(hash, "correct horse battery staple")
	if err != nil || !valid {
		t.Fatalf("valid=%v err=%v", valid, err)
	}
	valid, err = VerifyPassword(hash, "wrong password")
	if err != nil || valid {
		t.Fatalf("valid=%v err=%v", valid, err)
	}
}

func TestOpaqueRefreshTokenParsing(t *testing.T) {
	id := bson.NewObjectID()
	parsed, secret, err := parseToken(id.Hex() + ".secret")
	if err != nil || parsed != id || secret != "secret" {
		t.Fatalf("parsed=%v secret=%q err=%v", parsed, secret, err)
	}
	if _, _, err := parseToken("not-a-token"); err == nil {
		t.Fatal("invalid token accepted")
	}
}
