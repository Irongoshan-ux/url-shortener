package auth

type contextKey string

const (
	ContextKeyUserID            contextKey = "user_id"
	ContextKeyHadValidCookie    contextKey = "had_valid_cookie"
	ContextKeyHadCookieInRequest contextKey = "had_cookie_in_request"
)
