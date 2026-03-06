package nat

import "time"

type Table struct {
	byFlow   map[FlowKey]*Mapping
	byReturn map[ReverseKey]FlowKey
	now      func() time.Time
}

func NewTable(now func() time.Time) *Table {
	if now == nil {
		now = time.Now
	}
	return &Table{
		byFlow:   map[FlowKey]*Mapping{},
		byReturn: map[ReverseKey]FlowKey{},
		now:      now,
	}
}

func (t *Table) Upsert(flow FlowKey, translatedSrcPort uint16) *Mapping {
	if m, ok := t.byFlow[flow]; ok {
		m.LastSeen = t.now()
		return m
	}
	m := &Mapping{
		Original:          flow,
		TranslatedSrcIP:   flow.DstIP,
		TranslatedSrcPort: translatedSrcPort,
		LastSeen:          t.now(),
	}
	t.byFlow[flow] = m
	t.byReturn[ReverseKey{
		Proto:             flow.Proto,
		TranslatedDstIP:   m.TranslatedSrcIP,
		TranslatedDstPort: translatedSrcPort,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	}] = flow
	return m
}

func (t *Table) GetByFlow(flow FlowKey) (*Mapping, bool) {
	m, ok := t.byFlow[flow]
	return m, ok
}

func (t *Table) GetByReverse(key ReverseKey) (*Mapping, bool) {
	flow, ok := t.byReturn[key]
	if !ok {
		return nil, false
	}
	m, ok := t.byFlow[flow]
	return m, ok
}

func (t *Table) DeleteByFlow(flow FlowKey) {
	m, ok := t.byFlow[flow]
	if !ok {
		return
	}
	delete(t.byReturn, ReverseKey{
		Proto:             flow.Proto,
		TranslatedDstIP:   m.TranslatedSrcIP,
		TranslatedDstPort: m.TranslatedSrcPort,
		RemoteIP:          flow.DstIP,
		RemotePort:        flow.DstPort,
	})
	delete(t.byFlow, flow)
}
