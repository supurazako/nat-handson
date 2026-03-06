package forwarder

type ReverseLookup interface {
	Lookup(dstPort uint16) (clientPort uint16, ok bool)
}

// TODO: 返信パケットの宛先portを clientPort に戻す処理を実装する。
func TranslateInbound(pkt []byte, lookup ReverseLookup) (translated []byte, dropped bool, err error) {
	_ = pkt
	_ = lookup
	return nil, true, nil
}
