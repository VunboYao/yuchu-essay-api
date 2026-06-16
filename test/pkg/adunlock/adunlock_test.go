package adunlock_test

import (
	"testing"
	"time"

	"course_server/internal/pkg/adunlock"

	"github.com/golang-jwt/jwt/v5"
)

func TestSignAndVerify(t *testing.T) {
	token, expiresIn, err := adunlock.Sign("test-secret", "essay-mvp", 42, 7, "download_rewarded_video")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if expiresIn != 300 {
		t.Fatalf("expiresIn = %d, want 300", expiresIn)
	}
	claims, err := adunlock.Verify("test-secret", token, 42)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if claims.UserID != 7 || claims.SlotKey != "download_rewarded_video" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestVerifyFileMismatch(t *testing.T) {
	token, _, err := adunlock.Sign("test-secret", "essay-mvp", 42, 0, "download_rewarded_video")
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if _, err := adunlock.Verify("test-secret", token, 99); err == nil {
		t.Fatal("expected file mismatch error")
	}
}

func TestVerifyExpired(t *testing.T) {
	claims := adunlock.Claims{
		FileID:  1,
		SlotKey: "download_rewarded_video",
		Type:    "ad_unlock",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign expired: %v", err)
	}
	if _, err := adunlock.Verify("test-secret", token, 1); err == nil {
		t.Fatal("expected expired token error")
	}
}
