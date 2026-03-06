package nat

import "testing"

func TestAcquireExhaustionAndReuse(t *testing.T) {
	a := NewPortAllocator(40000, 40001)
	p1, err := a.Acquire()
	if err != nil {
		t.Fatalf("Acquire #1 failed: %v", err)
	}
	p2, err := a.Acquire()
	if err != nil {
		t.Fatalf("Acquire #2 failed: %v", err)
	}
	if p1 == p2 {
		t.Fatalf("expected unique ports: %d, %d", p1, p2)
	}
	if _, err := a.Acquire(); err == nil {
		t.Fatalf("expected exhaustion error")
	}
	a.Release(p1)
	if _, err := a.Acquire(); err != nil {
		t.Fatalf("expected acquire after release: %v", err)
	}
}
