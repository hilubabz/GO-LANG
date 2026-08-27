package main

import (
	"testing"

	"github.com/google/uuid"
	"github.com/hilubabz/GO-LANG/internal/auth"
)

func TestJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := auth.MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	gotID, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if gotID != userID {
		t.Errorf("expected %v, got %v", userID, gotID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := auth.MakeJWT(userID, secret)
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = auth.ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestWrongSecret(t *testing.T) {
	userID := uuid.New()

	token, err := auth.MakeJWT(userID, "correct-secret")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = auth.ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected wrong secret to be rejected")
	}
}