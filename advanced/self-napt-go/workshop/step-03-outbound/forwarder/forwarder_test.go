package forwarder

import "testing"

type fixedAlloc struct {
	p uint16
}

func (f fixedAlloc) Acquire() (uint16, error) { return f.p, nil }

func TestTranslateOutboundRewritesPacket(t *testing.T) {
	pkt := []byte{0x45, 0x00, 0x00, 0x28}
	out, err := TranslateOutbound(pkt, fixedAlloc{p: 40000})
	if err != nil {
		t.Fatalf("TranslateOutbound error: %v", err)
	}
	if out == nil {
		t.Fatalf("expected translated packet, got nil")
	}
}
