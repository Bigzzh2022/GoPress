package auth

import (
	"testing"
	"time"

	"gopress/internal/model"
)

func TestTokenRoundTripAndValidation(t *testing.T) {
	manager := NewTokenManager("a-sufficiently-long-test-secret", time.Minute)
	user := model.User{ID: 42, Role: model.RoleAdmin}
	value, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	identity, err := manager.Parse(value)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if identity.UserID != user.ID || identity.Role != user.Role {
		t.Fatalf("Parse() = (%d, %q), want (%d, %q)", identity.UserID, identity.Role, user.ID, user.Role)
	}
	if identity.JTI == "" {
		t.Fatal("Parse() returned an empty token id, so the token cannot be revoked")
	}
	if _, err := NewTokenManager("different-secret", time.Minute).Parse(value); err == nil {
		t.Fatal("Parse() accepted a token signed with another secret")
	}
}

func TestTokenIDsAreUnique(t *testing.T) {
	manager := NewTokenManager("a-sufficiently-long-test-secret", time.Minute)
	user := model.User{ID: 1, Role: model.RoleAdmin}
	first, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	second, err := manager.Generate(user)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	firstIdentity, err := manager.Parse(first)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	secondIdentity, err := manager.Parse(second)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if firstIdentity.JTI == secondIdentity.JTI {
		t.Fatal("two tokens shared a token id, so revoking one would revoke the other")
	}
}

func TestTokenRejectsExpiration(t *testing.T) {
	manager := NewTokenManager("a-sufficiently-long-test-secret", -time.Minute)
	value, err := manager.Generate(model.User{ID: 1, Role: model.RoleAdmin})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, err := manager.Parse(value); err == nil {
		t.Fatal("Parse() accepted an expired token")
	}
}
