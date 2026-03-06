package nat

import (
	"net/netip"
	"sync"
	"time"
)

type nowFunc func() time.Time

type Table struct {
	mu       sync.RWMutex
	now      nowFunc
	byFlow   map[FlowKey]*Mapping
	byReturn map[ReverseKey]FlowKey
}

func NewTable(now nowFunc) *Table {
	if now == nil {
		now = time.Now
	}
	return &Table{
		now:      now,
		byFlow:   make(map[FlowKey]*Mapping),
		byReturn: make(map[ReverseKey]FlowKey),
	}
}

func (t *Table) GetByFlow(flow FlowKey) (*Mapping, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	m, ok := t.byFlow[flow]
	if !ok {
		return nil, false
	}
	cp := *m
	return &cp, true
}

func (t *Table) Upsert(flow FlowKey, translatedSrcIP netip.Addr, translatedSrcPort uint16, state FlowState) *Mapping {
	t.mu.Lock()
	defer t.mu.Unlock()
	if m, ok := t.byFlow[flow]; ok {
		m.LastSeen = t.now()
		m.State = state
		cp := *m
		return &cp
	}
	m := &Mapping{
		Original:          flow,
		TranslatedSrcIP:   translatedSrcIP,
		TranslatedSrcPort: translatedSrcPort,
		LastSeen:          t.now(),
		State:             state,
	}
	t.byFlow[flow] = m
	retKey := ReverseKey{
		Proto:             flow.Proto,
		TranslatedDstIP:   translatedSrcIP,
		TranslatedDstPort: translatedSrcPort,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	}
	t.byReturn[retKey] = flow
	cp := *m
	return &cp
}

func (t *Table) GetByReverse(ret ReverseKey) (*Mapping, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	flow, ok := t.byReturn[ret]
	if !ok {
		return nil, false
	}
	m, ok := t.byFlow[flow]
	if !ok {
		return nil, false
	}
	m.LastSeen = t.now()
	cp := *m
	return &cp, true
}

func (t *Table) DeleteByFlow(flow FlowKey) {
	t.mu.Lock()
	defer t.mu.Unlock()
	m, ok := t.byFlow[flow]
	if !ok {
		return
	}
	retKey := ReverseKey{
		Proto:             flow.Proto,
		TranslatedDstIP:   m.TranslatedSrcIP,
		TranslatedDstPort: m.TranslatedSrcPort,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	}
	delete(t.byReturn, retKey)
	delete(t.byFlow, flow)
}

func (t *Table) Sweep(synTimeout, establishedTimeout time.Duration) []Mapping {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	var expired []Mapping
	for flow, m := range t.byFlow {
		timeout := establishedTimeout
		if m.State == StateSYNSent {
			timeout = synTimeout
		}
		if now.Sub(m.LastSeen) < timeout {
			continue
		}
		retKey := ReverseKey{
			Proto:             flow.Proto,
			TranslatedDstIP:   m.TranslatedSrcIP,
			TranslatedDstPort: m.TranslatedSrcPort,
			RemoteIP:          flow.DstIP,
			RemotePort:        flow.DstPort,
		}
		delete(t.byReturn, retKey)
		delete(t.byFlow, flow)
		expired = append(expired, *m)
	}
	return expired
}
