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
	id, role, err := manager.Parse(value)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if id != user.ID || role != user.Role {
		t.Fatalf("Parse() = (%d, %q), want (%d, %q)", id, role, user.ID, user.Role)
	}
	if _, _, err := NewTokenManager("different-secret", time.Minute).Parse(value); err == nil {
		t.Fatal("Parse() accepted a token signed with another secret")
	}
}

func TestTokenRejectsExpiration(t *testing.T) {
	manager := NewTokenManager("test-secret", -time.Minute)
	value, err := manager.Generate(model.User{ID: 1, Role: model.RoleAdmin})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if _, _, err := manager.Parse(value); err == nil {
		t.Fatal("Parse() accepted an expired token")
	}
}
