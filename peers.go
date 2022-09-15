package cache

// PeerGetter 查询缓存器，用于从对应的group获取缓存值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}

// PeerPicker 获取对应的查询缓存器
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}
