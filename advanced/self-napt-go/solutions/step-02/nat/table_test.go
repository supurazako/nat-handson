package nat

import (
	"net/netip"
	"testing"
	"time"
)

func TestUpsertAndReverseLookup(t *testing.T) {
	now := time.Unix(100, 0)
	tbl := NewTable(func() time.Time { return now })
	flow := FlowKey{
		Proto:   6,
		SrcIP:   netip.MustParseAddr("192.168.20.2"),
		SrcPort: 12345,
		DstIP:   netip.MustParseAddr("172.31.10.2"),
		DstPort: 80,
	}
	m := tbl.Upsert(flow, 40000)
	if m == nil {
		t.Fatalf("Upsert returned nil mapping")
	}
	rk := ReverseKey{
		Proto:             6,
		TranslatedDstIP:   m.TranslatedSrcIP,
		TranslatedDstPort: 40000,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	}
	if _, ok := tbl.GetByReverse(rk); !ok {
		t.Fatalf("expected reverse lookup hit")
	}
}

func TestDeleteByFlow(t *testing.T) {
	now := time.Unix(100, 0)
	tbl := NewTable(func() time.Time { return now })
	flow := FlowKey{
		Proto:   6,
		SrcIP:   netip.MustParseAddr("192.168.20.2"),
		SrcPort: 12345,
		DstIP:   netip.MustParseAddr("172.31.10.2"),
		DstPort: 80,
	}
	m := tbl.Upsert(flow, 40000)
	rk := ReverseKey{
		Proto:             6,
		TranslatedDstIP:   m.TranslatedSrcIP,
		TranslatedDstPort: 40000,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	}
	tbl.DeleteByFlow(flow)
	if _, ok := tbl.GetByFlow(flow); ok {
		t.Fatalf("expected flow deleted")
	}
	if _, ok := tbl.GetByReverse(rk); ok {
		t.Fatalf("expected reverse mapping deleted")
	}
}
