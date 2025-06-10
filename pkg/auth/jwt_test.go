package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateAndParseToken(t *testing.T) {
	username := "testuser"
	token, err := CreateToken(username, time.Minute)
	if err != nil {
		t.Fatalf("Failed to create token: %v", err)
	}

	claims, err := ParseToken(token)
	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if claims.Username != username {
		t.Errorf("Expected username %q, got %q", username, claims.Username)
	}
}

func TestParseToken_InvalidToken(t *testing.T) {
	invalidToken := "not.a.real.token"
	_, err := ParseToken(invalidToken)
	if err == nil {
		t.Error("Expected error for invalid token, got none")
	}
}

func TestExtractBearerToken(t *testing.T) {
	header := "Bearer sometoken123"
	token := extractBearerToken(header)
	if token != "sometoken123" {
		t.Errorf("Expected 'sometoken123', got %q", token)
	}

	header = "Basic abc"
	token = extractBearerToken(header)
	if token != "" {
		t.Errorf("Expected '', got %q", token)
	}
}

func TestMiddleware_Success(t *testing.T) {
	token, _ := CreateToken("user1", time.Minute)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()

	var gotUser string

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := FromContext(r.Context())
		if !ok {
			t.Fatal("No claims in context")
		}
		gotUser = claims.Username
		w.WriteHeader(http.StatusOK)
	}))

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if gotUser != "user1" {
		t.Errorf("Expected username 'user1', got %q", gotUser)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rr := httptest.NewRecorder()

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called without a token")
	}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing header, got %d", rr.Code)
	}
}

func TestMiddleware_InvalidToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	rr := httptest.NewRecorder()

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("Handler should not be called with invalid token")
	}))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", rr.Code)
	}
}

func TestFromContext(t *testing.T) {
	claims := &Claims{Username: "ctxuser"}
	ctx := context.WithValue(context.Background(), ContextUserKeyKey, claims)

	gotClaims, ok := FromContext(ctx)
	if !ok {
		t.Fatal("Expected claims in context")
	}
	if gotClaims.Username != "ctxuser" {
		t.Errorf("Expected username 'ctxuser', got %q", gotClaims.Username)
	}
}
