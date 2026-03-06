package forwarder

import (
	"encoding/binary"
	"testing"
)

type fakeLookup struct {
	ok bool
}

func (f fakeLookup) Lookup(dstPort uint16) (uint16, bool) {
	_ = dstPort
	if !f.ok {
		return 0, false
	}
	return 12345, true
}

func TestTranslateInboundTracked(t *testing.T) {
	pkt := []byte{0x45, 0x00, 0x00, 0x50}
	out, dropped, err := TranslateInbound(pkt, fakeLookup{ok: true})
	if err != nil {
		t.Fatalf("TranslateInbound error: %v", err)
	}
	if dropped {
		t.Fatalf("expected tracked packet not dropped")
	}
	if out == nil {
		t.Fatalf("expected translated packet")
	}
	if got := binary.BigEndian.Uint16(out[len(out)-2:]); got != 12345 {
		t.Fatalf("expected restored client port=12345, got %d", got)
	}
}

func TestTranslateInboundUntracked(t *testing.T) {
	pkt := []byte{0x45, 0x00, 0x00, 0x50}
	_, dropped, err := TranslateInbound(pkt, fakeLookup{ok: false})
	if err != nil {
		t.Fatalf("TranslateInbound error: %v", err)
	}
	if !dropped {
		t.Fatalf("expected untracked packet dropped")
	}
}
