package service

// Option configures Service construction (functional options pattern).
type Option func(*Service)

// WithMaxAttempts sets how many times ShortenURL retries on short-id collision before failing.
func WithMaxAttempts(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.maxAttempts = n
		}
	}
}

// WithIDGenerator replaces the default random base64 short id generator (e.g. for deterministic tests).
func WithIDGenerator(fn func() (string, error)) Option {
	return func(s *Service) {
		if fn != nil {
			s.idGen = fn
		}
	}
}
