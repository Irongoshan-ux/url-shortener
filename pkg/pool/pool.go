package pool

import (
	"reflect"
	"sync"
)

type Resetter interface {
	Reset()
}

type Pool[T Resetter] struct {
	p sync.Pool
}

func New[T Resetter]() *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any {
				return newT[T]()
			},
		},
	}
}

func (pool *Pool[T]) Get() T {
	v := pool.p.Get()
	if v == nil {
		return newT[T]()
	}
	return v.(T)
}

func (pool *Pool[T]) Put(x T) {
	x.Reset()
	pool.p.Put(x)
}

func newT[T Resetter]() T {
	var zero T
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Pointer {
		return reflect.New(t.Elem()).Interface().(T)
	}
	return zero
}
