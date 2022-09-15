package singleflight

import "sync"

// call 请求实例
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// Group 防止并发实体
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 确保方法只执行一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	// 1.加锁避免并发
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 2.如果map中有值，意味着有请求进行中
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err
	}

	// 3.没有创建一个到map中，代表任务正在进行中
	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	// 4.执行任务
	c.val, c.err = fn()
	c.wg.Done()

	// 5.删除字典，完成请求
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
