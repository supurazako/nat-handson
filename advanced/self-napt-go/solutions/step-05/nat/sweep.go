package nat

import "time"

type State int

const (
	StateSYNSent State = iota
	StateEstablished
)

type Mapping struct {
	State    State
	LastSeen time.Time
}

func Sweep(now time.Time, in []Mapping, synTimeout, establishedTimeout time.Duration) (alive []Mapping, expired []Mapping) {
	for _, m := range in {
		timeout := establishedTimeout
		if m.State == StateSYNSent {
			timeout = synTimeout
		}
		if now.Sub(m.LastSeen) >= timeout {
			expired = append(expired, m)
			continue
		}
		alive = append(alive, m)
	}
	return alive, expired
}
