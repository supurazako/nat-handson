package nat

import (
	"net/netip"
	"testing"
	"time"
)

func TestTableUpsertAndReverseLookup(t *testing.T) {
	now := time.Unix(100, 0)
	tbl := NewTable(func() time.Time { return now })
	flow := FlowKey{
		Proto:   ProtoTCP,
		SrcIP:   netip.MustParseAddr("192.168.10.2"),
		SrcPort: 12345,
		DstIP:   netip.MustParseAddr("172.31.10.2"),
		DstPort: 80,
	}
	m := tbl.Upsert(flow, netip.MustParseAddr("172.31.10.1"), 40000, StateSYNSent)
	if m.TranslatedSrcPort != 40000 {
		t.Fatalf("unexpected translated port: %d", m.TranslatedSrcPort)
	}

	ret := ReverseKey{
		Proto:             ProtoTCP,
		TranslatedDstIP:   netip.MustParseAddr("172.31.10.1"),
		TranslatedDstPort: 40000,
		RemoteIP:          netip.MustParseAddr("172.31.10.2"),
		RemotePort:        80,
	}
	if _, ok := tbl.GetByReverse(ret); !ok {
		t.Fatalf("expected reverse lookup hit")
	}
}

func TestTableSweep(t *testing.T) {
	now := time.Unix(200, 0)
	tbl := NewTable(func() time.Time { return now })
	flow := FlowKey{
		Proto:   ProtoTCP,
		SrcIP:   netip.MustParseAddr("192.168.10.2"),
		SrcPort: 20000,
		DstIP:   netip.MustParseAddr("172.31.10.2"),
		DstPort: 80,
	}
	tbl.Upsert(flow, netip.MustParseAddr("172.31.10.1"), 40001, StateSYNSent)
	now = now.Add(31 * time.Second)
	expired := tbl.Sweep(30*time.Second, 5*time.Minute)
	if len(expired) != 1 {
		t.Fatalf("expected 1 expired mapping, got %d", len(expired))
	}
}
