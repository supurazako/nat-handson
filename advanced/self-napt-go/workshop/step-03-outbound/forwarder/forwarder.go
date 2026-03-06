package forwarder

type Allocator interface {
	Acquire() (uint16, error)
}

// TODO: packet parse/serialize を実装して送信元portを書き換える。
func TranslateOutbound(pkt []byte, alloc Allocator) ([]byte, error) {
	_ = pkt
	_ = alloc
	return nil, nil
}
