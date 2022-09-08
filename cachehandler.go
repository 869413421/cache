package cache

import (
	"fmt"
	"log"
	"sync"
)

// Getter 获取数据回调接口
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 实现Getter接口的方法类型
type GetterFunc func(key string) ([]byte, error)

// Get 获取缓存数据
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// Group 缓存分组实体
type Group struct {
	name   string // 分组名称
	getter Getter // 缓存获取实现
	cache  cache  // 缓存实体
}

// NewGroup 创建新的缓存分组
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:   name,
		getter: getter,
		cache:  cache{cacheBytes: cacheBytes},
	}

	groups[name] = g
	return g
}

// GetGroup 获取缓存分组
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// Get 获取缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.cache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// load 载入缓存，返回数据
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
}

// getLocally 从getter获取数据然后加载到缓存当中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// populateCache 保存值的缓存中
func (g *Group) populateCache(key string, value ByteView) {
	g.cache.add(key, value)
}
