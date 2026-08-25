package service

import (
	"testing"

	"github.com/remiehneppo/material-management/types"
)

func TestAccessTokenPreservesClaimsAndSessionID(t *testing.T) {
	jwtService := NewJWTService("test-secret", "test-issuer", 60)
	want := &types.User{ID: "user-id", Username: "tester", Workspace: "dock", WorkspaceRole: "staff"}
	token, err := jwtService.GenerateAccessTokenForSession(want, "session-id")
	if err != nil {
		t.Fatal(err)
	}
	got, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != want.ID || got.Username != want.Username || got.Workspace != want.Workspace || got.WorkspaceRole != want.WorkspaceRole || got.SessionID != "session-id" {
		t.Fatalf("claims=%+v", got)
	}
}
