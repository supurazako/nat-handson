package forwarder

import "testing"

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
	pkt := []byte{0x45, 0x00}
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
}

func TestTranslateInboundUntracked(t *testing.T) {
	pkt := []byte{0x45, 0x00}
	_, dropped, err := TranslateInbound(pkt, fakeLookup{ok: false})
	if err != nil {
		t.Fatalf("TranslateInbound error: %v", err)
	}
	if !dropped {
		t.Fatalf("expected untracked packet dropped")
	}
}
