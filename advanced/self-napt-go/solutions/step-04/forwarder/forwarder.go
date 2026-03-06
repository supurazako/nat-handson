package forwarder

import "encoding/binary"

type ReverseLookup interface {
	Lookup(dstPort uint16) (clientPort uint16, ok bool)
}

func TranslateInbound(pkt []byte, lookup ReverseLookup) (translated []byte, dropped bool, err error) {
	if len(pkt) < 2 {
		return append([]byte(nil), pkt...), false, nil
	}
	dstPort := binary.BigEndian.Uint16(pkt[len(pkt)-2:])
	clientPort, ok := lookup.Lookup(dstPort)
	if !ok {
		return nil, true, nil
	}
	out := append([]byte(nil), pkt...)
	binary.BigEndian.PutUint16(out[len(out)-2:], clientPort)
	return out, false, nil
}
