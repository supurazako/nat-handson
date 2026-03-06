package packet

import (
	"encoding/binary"
	"fmt"
)

const tcpHeaderMinLen = 20

type TCPHeader struct {
	SrcPort uint16
	DstPort uint16
	Flags   uint8
}

func ParseTCPHeader(pkt []byte) (TCPHeader, error) {
	if len(pkt) < ipv4HeaderLen+tcpHeaderMinLen {
		return TCPHeader{}, fmt.Errorf("packet too short for tcp")
	}
	off := ipv4HeaderLen
	return TCPHeader{
		SrcPort: binary.BigEndian.Uint16(pkt[off : off+2]),
		DstPort: binary.BigEndian.Uint16(pkt[off+2 : off+4]),
		Flags:   pkt[off+13],
	}, nil
}

func RewriteTCPPorts(pkt []byte, srcPort, dstPort uint16) {
	off := ipv4HeaderLen
	binary.BigEndian.PutUint16(pkt[off:off+2], srcPort)
	binary.BigEndian.PutUint16(pkt[off+2:off+4], dstPort)
}

func IsSYN(flags uint8) bool {
	return flags&0x02 != 0 && flags&0x10 == 0
}
