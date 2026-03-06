package nat

import (
	"fmt"
	"net/netip"
	"time"
)

type Proto uint8

const (
	ProtoTCP Proto = 6
)

type FlowState int

const (
	StateSYNSent FlowState = iota
	StateEstablished
	StateFINWait
	StateTimeWait
)

func (s FlowState) String() string {
	switch s {
	case StateSYNSent:
		return "SYN_SENT"
	case StateEstablished:
		return "ESTABLISHED"
	case StateFINWait:
		return "FIN_WAIT"
	case StateTimeWait:
		return "TIME_WAIT"
	default:
		return "UNKNOWN"
	}
}

type FlowKey struct {
	Proto   Proto
	SrcIP   netip.Addr
	SrcPort uint16
	DstIP   netip.Addr
	DstPort uint16
}

func (k FlowKey) String() string {
	return fmt.Sprintf("%d:%s:%d->%s:%d", k.Proto, k.SrcIP, k.SrcPort, k.DstIP, k.DstPort)
}

type ReverseKey struct {
	Proto             Proto
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
	State             FlowState
}
