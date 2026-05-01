package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const CookieName = "user_id"

const cookieMaxAge = 3600 * 24 * 30 // 30 days

var ErrInvalidCookie = errors.New("invalid cookie")

type claims struct {
	jwt.RegisteredClaims
	UserID string `json:"uid"`
}

func makeCookie(secret, userID string) (*http.Cookie, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cookieMaxAge) * time.Second)),
		},
		UserID: userID,
	})
	signed, err := token.SignedString(secretBytes(secret))
	if err != nil {
		return nil, err
	}
	return &http.Cookie{
		Name:     CookieName,
		Value:    signed,
		Path:     "/",
		MaxAge:   cookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func secretBytes(s string) []byte {
	if s == "" {
		return []byte("default-secret")
	}
	return []byte(s)
}

func verifyAndGetUserID(secret, cookieValue string) (string, error) {
	token, err := jwt.ParseWithClaims(cookieValue, &claims{}, func(_ *jwt.Token) (interface{}, error) {
		return secretBytes(secret), nil
	})
	if err != nil {
		return "", ErrInvalidCookie
	}
	c, ok := token.Claims.(*claims)
	if !ok || !token.Valid || c.UserID == "" {
		return "", ErrInvalidCookie
	}
	return c.UserID, nil
}

func GenerateUserID() string {
	return uuid.New().String()
}

func GetOrCreateUserID(r *http.Request, secret string) (userID string, hadValidCookie bool, hadCookieInRequest bool, setCookie *http.Cookie) {
	cookie, err := r.Cookie(CookieName)
	if err != nil || cookie == nil || cookie.Value == "" {
		userID = GenerateUserID()
		c, err := makeCookie(secret, userID)
		if err != nil {
			return userID, false, false, nil
		}
		return userID, false, false, c
	}
	userID, err = verifyAndGetUserID(secret, cookie.Value)
	if err != nil {
		userID = GenerateUserID()
		c, err := makeCookie(secret, userID)
		if err != nil {
			return userID, false, true, nil
		}
		return userID, false, true, c
	}
	return userID, true, true, nil
}
