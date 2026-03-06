package nat

import "testing"

func TestPortAllocatorAcquireRelease(t *testing.T) {
	a := NewPortAllocator(40000, 40001)
	p1, err := a.Acquire()
	if err != nil {
		t.Fatalf("Acquire 1 failed: %v", err)
	}
	p2, err := a.Acquire()
	if err != nil {
		t.Fatalf("Acquire 2 failed: %v", err)
	}
	if p1 == p2 {
		t.Fatalf("expected different ports, got %d", p1)
	}
	if _, err := a.Acquire(); err == nil {
		t.Fatalf("expected exhaustion error")
	}
	a.Release(p1)
	if _, err := a.Acquire(); err != nil {
		t.Fatalf("expected port after release: %v", err)
	}
}
