package lru

import "container/list"

type (
	// Value 内存存储接口，实现Len接口，len返回当前实体所占内存大小
	Value interface {
		Len() int
	}

	// entry Cache中ll存储的实际元素
	entry struct {
		key   string
		value Value
	}

	Cache struct {
		maxBytes  int64                         // 允许的最大内存
		nBytes    int64                         // 已经使用的内存
		ll        *list.List                    // 双向链表，存储数据
		cache     map[string]*list.Element      // 对应字典，key对应双向链表数据的指针
		OnEvicted func(key string, value Value) // 某个key被移除时的回调函数，允许为nil
	}
)

// New 创建返回缓存对象
func New(maxBytes int64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		OnEvicted: onEvicted,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
	}
}

// Get 获取缓存
func (c *Cache) Get(key string) (value Value, ok bool) {
	// 1.查找对应key数据是否存在
	element, ok := c.cache[key]
	if ok {
		// 1.1 将元素放到队尾
		c.ll.MoveToFront(element)
		// 1.2 断言转换为entry
		kv := element.Value.(*entry)
		// 1.3 返回存储的实际值
		return kv.value, true
	}
	return
}

// RemoveOldest 淘汰数据
func (c *Cache) RemoveOldest() {
	// 1.取队列首个元素
	element := c.ll.Back()
	if element == nil {
		return
	}

	// 2.从队列中删除删除元素
	c.ll.Remove(element)

	// 3.删除字典对应的key
	e := element.Value.(*entry)
	delete(c.cache, e.key)

	// 4.重新计算已经使用的容量
	c.nBytes -= int64(len(e.key)) + int64(e.value.Len())

	// 5.如果有注册回调函数，执行
	if c.OnEvicted != nil {
		c.OnEvicted(e.key, e.value)
	}
}

func (c *Cache) Add(key string, value Value) {
	// 1.检查key是否已经存在
	if element, ok := c.cache[key]; ok {
		// 1.1.如果存在，将元素移动到队尾
		c.ll.MoveToFront(element)
		// 1.2.取出原来存储的值
		e := element.Value.(*entry)
		// 1.3.重新计算已使用容量
		c.nBytes += int64(value.Len()) - int64(e.value.Len())
		e.value = value
	} else {
		// 1.4.若果数据不存在，往队尾新增节点
		element = c.ll.PushFront(&entry{key: key, value: value})
		// 1.5 存储对应的指针
		c.cache[key] = element
		// 1.6 计算容量
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	// 2.淘汰缓存，知道容量到达允许的最大容量
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// Len 缓存数量
func (c *Cache) Len() int {
	return c.ll.Len()
}
