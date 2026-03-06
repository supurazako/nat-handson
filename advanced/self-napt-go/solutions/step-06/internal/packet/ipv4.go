package packet

import (
	"encoding/binary"
	"fmt"
	"net/netip"
)

const ipv4HeaderLen = 20

type IPv4Header struct {
	Protocol uint8
	Src      netip.Addr
	Dst      netip.Addr
	TTL      uint8
	TotalLen uint16
}

func ParseIPv4Header(pkt []byte) (IPv4Header, error) {
	if len(pkt) < ipv4HeaderLen {
		return IPv4Header{}, fmt.Errorf("packet too short")
	}
	ihl := pkt[0] & 0x0f
	if ihl != 5 {
		return IPv4Header{}, fmt.Errorf("ipv4 options unsupported")
	}
	src, ok := netip.AddrFromSlice(pkt[12:16])
	if !ok {
		return IPv4Header{}, fmt.Errorf("invalid source ip")
	}
	dst, ok := netip.AddrFromSlice(pkt[16:20])
	if !ok {
		return IPv4Header{}, fmt.Errorf("invalid destination ip")
	}
	return IPv4Header{
		Protocol: pkt[9],
		Src:      src,
		Dst:      dst,
		TTL:      pkt[8],
		TotalLen: binary.BigEndian.Uint16(pkt[2:4]),
	}, nil
}

func RewriteIPv4Endpoints(pkt []byte, src, dst netip.Addr) error {
	src4 := src.As4()
	dst4 := dst.As4()
	copy(pkt[12:16], src4[:])
	copy(pkt[16:20], dst4[:])
	// Clear then recompute header checksum.
	pkt[10] = 0
	pkt[11] = 0
	sum := Checksum(pkt[:ipv4HeaderLen])
	binary.BigEndian.PutUint16(pkt[10:12], sum)
	return nil
}
