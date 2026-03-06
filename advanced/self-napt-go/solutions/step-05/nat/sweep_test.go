package nat

import (
	"testing"
	"time"
)

func TestSweepByState(t *testing.T) {
	now := time.Unix(1000, 0)
	in := []Mapping{
		{State: StateSYNSent, LastSeen: now.Add(-31 * time.Second)},
		{State: StateEstablished, LastSeen: now.Add(-31 * time.Second)},
	}
	alive, expired := Sweep(now, in, 30*time.Second, 5*time.Minute)
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired mapping, got %d", len(expired))
	}
	if len(alive) != 1 {
		t.Fatalf("expected 1 alive mapping, got %d", len(alive))
	}
}
