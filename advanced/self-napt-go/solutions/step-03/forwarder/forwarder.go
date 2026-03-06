package forwarder

import "encoding/binary"

type Allocator interface {
	Acquire() (uint16, error)
}

func TranslateOutbound(pkt []byte, alloc Allocator) ([]byte, error) {
	newPort, err := alloc.Acquire()
	if err != nil {
		return nil, err
	}
	out := append([]byte(nil), pkt...)
	if len(out) < 2 {
		return out, nil
	}
	binary.BigEndian.PutUint16(out[len(out)-2:], newPort)
	return out, nil
}
