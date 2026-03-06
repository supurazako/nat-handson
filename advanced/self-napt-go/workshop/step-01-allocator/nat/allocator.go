package nat

import "fmt"

type PortAllocator struct {
	min   uint16
	max   uint16
	next  uint16
	inUse map[uint16]struct{}
}

func NewPortAllocator(min, max uint16) *PortAllocator {
	return &PortAllocator{
		min:   min,
		max:   max,
		next:  min,
		inUse: map[uint16]struct{}{},
	}
}

func (a *PortAllocator) Acquire() (uint16, error) {
	// TODO: 実装する
	return 0, fmt.Errorf("TODO: implement Acquire")
}

func (a *PortAllocator) Release(port uint16) {
	// TODO: 実装する
	_ = port
}
