package forwarder

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/netip"
	"time"

	"nat-handson/advanced/self-napt-go/prototype/internal/nat"
	"nat-handson/advanced/self-napt-go/prototype/internal/packet"
)

type portAllocator interface {
	Acquire() (uint16, error)
	Release(port uint16)
}

type Forwarder struct {
	table   *nat.Table
	alloc   portAllocator
	wanIP   netip.Addr
	logger  *log.Logger
	closers []io.Closer
}

func New(table *nat.Table, alloc portAllocator, wanIP net.IP, logger *log.Logger) *Forwarder {
	wanAddr, _ := netip.AddrFromSlice(wanIP.To4())
	return &Forwarder{
		table:  table,
		alloc:  alloc,
		wanIP:  wanAddr,
		logger: logger,
	}
}

// TranslateOutbound mutates LAN->WAN packets in-place.
func (f *Forwarder) TranslateOutbound(pkt []byte) error {
	ipv4h, err := packet.ParseIPv4Header(pkt)
	if err != nil {
		return err
	}
	if ipv4h.Protocol != uint8(nat.ProtoTCP) {
		return nil
	}
	tcph, err := packet.ParseTCPHeader(pkt)
	if err != nil {
		return err
	}

	flow := nat.FlowKey{
		Proto:   nat.ProtoTCP,
		SrcIP:   ipv4h.Src,
		SrcPort: tcph.SrcPort,
		DstIP:   ipv4h.Dst,
		DstPort: tcph.DstPort,
	}

	var m *nat.Mapping
	if existing, ok := f.table.GetByFlow(flow); ok {
		m = existing
	} else {
		p, allocErr := f.alloc.Acquire()
		if allocErr != nil {
			return allocErr
		}
		state := nat.StateEstablished
		if packet.IsSYN(tcph.Flags) {
			state = nat.StateSYNSent
		}
		m = f.table.Upsert(flow, f.wanIP, p, state)
		f.logJSON("create_mapping", map[string]any{
			"flow":                flow.String(),
			"translated_src_ip":   m.TranslatedSrcIP.String(),
			"translated_src_port": m.TranslatedSrcPort,
		})
	}

	if err := packet.RewriteIPv4Endpoints(pkt, m.TranslatedSrcIP, ipv4h.Dst); err != nil {
		return err
	}
	packet.RewriteTCPPorts(pkt, m.TranslatedSrcPort, tcph.DstPort)
	f.logJSON("translate_outbound", map[string]any{
		"orig_src_ip":   ipv4h.Src.String(),
		"orig_src_port": tcph.SrcPort,
		"new_src_ip":    m.TranslatedSrcIP.String(),
		"new_src_port":  m.TranslatedSrcPort,
		"dst_ip":        ipv4h.Dst.String(),
		"dst_port":      tcph.DstPort,
	})
	return nil
}

// TranslateInbound mutates WAN->LAN packets in-place.
func (f *Forwarder) TranslateInbound(pkt []byte) error {
	ipv4h, err := packet.ParseIPv4Header(pkt)
	if err != nil {
		return err
	}
	if ipv4h.Protocol != uint8(nat.ProtoTCP) {
		return nil
	}
	tcph, err := packet.ParseTCPHeader(pkt)
	if err != nil {
		return err
	}

	ret := nat.ReverseKey{
		Proto:             nat.ProtoTCP,
		TranslatedDstIP:   ipv4h.Dst,
		TranslatedDstPort: tcph.DstPort,
		RemoteIP:          ipv4h.Src,
		RemotePort:        tcph.SrcPort,
	}
	m, ok := f.table.GetByReverse(ret)
	if !ok {
		f.logJSON("drop_untracked_packet", map[string]any{
			"src_ip":   ipv4h.Src.String(),
			"src_port": tcph.SrcPort,
			"dst_ip":   ipv4h.Dst.String(),
			"dst_port": tcph.DstPort,
		})
		return nil
	}

	if err := packet.RewriteIPv4Endpoints(pkt, ipv4h.Src, m.Original.SrcIP); err != nil {
		return err
	}
	packet.RewriteTCPPorts(pkt, tcph.SrcPort, m.Original.SrcPort)
	f.logJSON("translate_inbound", map[string]any{
		"src_ip":       ipv4h.Src.String(),
		"src_port":     tcph.SrcPort,
		"orig_dst_ip":  ipv4h.Dst.String(),
		"orig_dst_port": tcph.DstPort,
		"new_dst_ip":   m.Original.SrcIP.String(),
		"new_dst_port": m.Original.SrcPort,
	})
	return nil
}

func (f *Forwarder) RegisterCloser(c io.Closer) {
	f.closers = append(f.closers, c)
}

func (f *Forwarder) Close() {
	for _, c := range f.closers {
		_ = c.Close()
	}
}

func (f *Forwarder) logJSON(event string, kv map[string]any) {
	record := map[string]any{
		"ts":    time.Now().UTC().Format(time.RFC3339),
		"event": event,
	}
	for k, v := range kv {
		record[k] = v
	}
	b, err := json.Marshal(record)
	if err != nil {
		f.logger.Printf("{\"event\":\"log_error\",\"message\":%q}", err.Error())
		return
	}
	f.logger.Println(string(b))
}
