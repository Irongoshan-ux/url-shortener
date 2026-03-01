package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const CookieName = "user_id"

var ErrInvalidCookie = errors.New("invalid cookie")

func SignUserID(secret, userID string) string {
	if secret == "" {
		secret = "default-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	sig := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	id := base64.URLEncoding.EncodeToString([]byte(userID))
	return id + "." + sig
}

func VerifyAndGetUserID(secret, cookieValue string) (string, error) {
	if secret == "" {
		secret = "default-secret"
	}
	parts := strings.SplitN(cookieValue, ".", 2)
	if len(parts) != 2 {
		return "", ErrInvalidCookie
	}
	idEnc, sigEnc := parts[0], parts[1]
	idBytes, err := base64.URLEncoding.DecodeString(idEnc)
	if err != nil {
		return "", ErrInvalidCookie
	}
	userID := string(idBytes)
	if userID == "" {
		return "", ErrInvalidCookie
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(userID))
	expected := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(sigEnc), []byte(expected)) {
		return "", ErrInvalidCookie
	}
	return userID, nil
}

func GenerateUserID() string {
	return uuid.New().String()
}

func GetOrCreateUserID(r *http.Request, secret string) (userID string, hadValidCookie bool, setCookie *http.Cookie) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie.Value == "" {
		userID = GenerateUserID()
		signed := SignUserID(secret, userID)
		setCookie = &http.Cookie{
			Name:  CookieName,
			Value: signed,
			Path:  "/",
			MaxAge: 3600 * 24 * 365,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		return userID, false, setCookie
	}
	userID, err = VerifyAndGetUserID(secret, cookie.Value)
	if err != nil {
		userID = GenerateUserID()
		signed := SignUserID(secret, userID)
		setCookie = &http.Cookie{
			Name:  CookieName,
			Value: signed,
			Path:  "/",
			MaxAge: 3600 * 24 * 365,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		}
		return userID, false, setCookie
	}
	return userID, true, nil
}
