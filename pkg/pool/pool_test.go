package pool_test

import (
	"testing"

	"github.com/Irongoshan-ux/url-shortener/pkg/pool"
)

type testItem struct {
	n int
}

func (t *testItem) Reset() {
	if t == nil {
		return
	}
	t.n = 0
}

func TestPool_PutResetsBeforeReuse(t *testing.T) {
	p := pool.New[*testItem]()

	x := p.Get()
	x.n = 42
	p.Put(x)

	y := p.Get()
	if y.n != 0 {
		t.Fatalf("after Put/Get: n=%d, want 0", y.n)
	}
}

func TestPool_GetReturnsDistinctObjectsWhenEmpty(t *testing.T) {
	p := pool.New[*testItem]()

	a := p.Get()
	b := p.Get()
	if a == b {
		t.Fatal("expected two distinct allocations from empty pool")
	}
	p.Put(a)
	p.Put(b)

	c := p.Get()
	d := p.Get()
	if c == d {
		t.Fatal("expected two objects from pool")
	}
}
