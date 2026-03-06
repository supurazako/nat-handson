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
		inUse: make(map[uint16]struct{}),
	}
}

func (a *PortAllocator) Acquire() (uint16, error) {
	span := int(a.max-a.min) + 1
	for i := 0; i < span; i++ {
		p := a.next
		a.next++
		if a.next > a.max {
			a.next = a.min
		}
		if _, exists := a.inUse[p]; exists {
			continue
		}
		a.inUse[p] = struct{}{}
		return p, nil
	}
	return 0, fmt.Errorf("port range exhausted")
}

func (a *PortAllocator) Reserve(port uint16) bool {
	if port < a.min || port > a.max {
		return false
	}
	if _, exists := a.inUse[port]; exists {
		return false
	}
	a.inUse[port] = struct{}{}
	return true
}

func (a *PortAllocator) Release(port uint16) {
	delete(a.inUse, port)
}
