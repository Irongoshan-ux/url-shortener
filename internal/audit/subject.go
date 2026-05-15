package audit

import (
	"context"
	"sync"
)

type Observer interface {
	OnAudit(ctx context.Context, e Event)
}

type Subject struct {
	mu        sync.RWMutex
	observers []Observer
}

func NewSubject(observers ...Observer) *Subject {
	s := &Subject{}
	for _, o := range observers {
		if o != nil {
			s.observers = append(s.observers, o)
		}
	}
	return s
}

func (s *Subject) Attach(o Observer) {
	if o == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.observers = append(s.observers, o)
}

func (s *Subject) Notify(ctx context.Context, e Event) {
	s.mu.RLock()
	list := append([]Observer(nil), s.observers...)
	s.mu.RUnlock()
	for _, o := range list {
		go o.OnAudit(ctx, e)
	}
}
