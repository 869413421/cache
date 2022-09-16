package cache

import (
	"fmt"
	pb "github.com/869413421/cache/proto/pbcache"
	"github.com/869413421/cache/singleflight"
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
	name   string              // 分组名称
	getter Getter              // 缓存获取实现
	cache  cache               // 缓存实体
	peers  PeerPicker          //缓存选择器
	loader *singleflight.Group // 加锁请求，避免缓存击穿
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
		loader: &singleflight.Group{},
	}

	groups[name] = g
	return g
}

// RegisterPeers 注册缓存选择器
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// getFormPeer 从远程节点获取缓存
func (g *Group) getFormPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{Group: g.name, Key: key}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}

// load 载入缓存，返回数据
func (g *Group) load(key string) (value ByteView, err error) {
	view, err := g.loader.Do(key, func() (interface{}, error) {
		// 1.如果有远程节点，首先判断是否从远程节点获取
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFormPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[Cache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return view.(ByteView), nil
	}

	return
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
	// 1.key不允许为空
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	// 2.检查是否本地已有缓存，有直接返回
	if v, ok := g.cache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	// 3.没有缓存策略加载缓存
	return g.load(key)
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
