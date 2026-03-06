package nat

import (
	"net/netip"
	"time"
)

type FlowKey struct {
	Proto   uint8
	SrcIP   netip.Addr
	SrcPort uint16
	DstIP   netip.Addr
	DstPort uint16
}

type ReverseKey struct {
	Proto             uint8
	TranslatedDstIP   netip.Addr
	TranslatedDstPort uint16
	RemoteIP          netip.Addr
	RemotePort        uint16
}

type Mapping struct {
	Original          FlowKey
	TranslatedSrcIP   netip.Addr
	TranslatedSrcPort uint16
	LastSeen          time.Time
}
