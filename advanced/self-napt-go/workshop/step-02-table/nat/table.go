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
	// TODO: 実装する
	_ = flow
	_ = translatedSrcPort
	return nil
}

func (t *Table) GetByFlow(flow FlowKey) (*Mapping, bool) {
	// TODO: 実装する
	_ = flow
	return nil, false
}

func (t *Table) GetByReverse(key ReverseKey) (*Mapping, bool) {
	// TODO: 実装する
	_ = key
	return nil, false
}

func (t *Table) DeleteByFlow(flow FlowKey) {
	// TODO: 実装する
	_ = flow
}
