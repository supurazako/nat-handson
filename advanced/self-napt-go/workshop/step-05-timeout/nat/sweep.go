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

// TODO: 状態ごとの timeout を使って期限切れ判定する。
func Sweep(now time.Time, in []Mapping, synTimeout, establishedTimeout time.Duration) (alive []Mapping, expired []Mapping) {
	_ = now
	_ = synTimeout
	_ = establishedTimeout
	return in, nil
}
