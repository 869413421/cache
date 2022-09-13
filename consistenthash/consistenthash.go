package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// Hash 自定义Hash方法，方便替换
type Hash func(data []byte) uint32

// Map 一致性哈希结构
type Map struct {
	hash     Hash           // hash算法
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点映射，key是虚拟节点哈希值，v是真实节点名称
}

// New 创建一致性哈希结构结构体
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add 新增虚拟节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get 获取真实节点名称
func (m *Map) Get(key string) string {
	// 1.哈希环为空，返回空节点
	if len(m.keys) == 0 {
		return ""
	}

	// 2.获取key的哈希值
	hash := int(m.hash([]byte(key)))

	// 3.获取第一个符合的hash的下标
	index := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//fmt.Printf("长度%d \n", len(m.keys))
	//fmt.Printf("下标%d \n", index)
	//fmt.Printf("取余 %d \n", m.keys[index%len(m.keys)])
	//fmt.Println(m.hashMap)
	// 4.获取第一个符合的hash的下标
	return m.hashMap[m.keys[index%len(m.keys)]]
}
