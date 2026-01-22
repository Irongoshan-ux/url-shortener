package service

type Option func(*Service)

func WithMaxAttempts(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.maxAttempts = n
		}
	}
}

func WithIDGenerator(fn func() (string, error)) Option {
	return func(s *Service) {
		if fn != nil {
			s.idGen = fn
		}
	}
}


