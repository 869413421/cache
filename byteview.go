package cache

type ByteView struct {
	b []byte
}

// cloneBytes 克隆字节供外部使用，防止被篡改
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// Len 实现Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// String 转换字符串
func (v ByteView) String() string {
	return string(v.b)
}

// ByteSlice 返回当前字节副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}
