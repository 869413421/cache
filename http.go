package cache

import (
	"fmt"
	"github.com/869413421/cache/consistenthash"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	// 默认访问前缀
	defaultBasePath = "/_cache/"

	//虚拟节点数
	defaultReplicas = 50
)

// HTTPPool 服务端连接
type HTTPPool struct {
	self        string // 服务端名称
	basePath    string // 服务端路径
	mu          sync.Mutex
	peers       *consistenthash.Map    // 节点哈希环
	httpGetters map[string]*HttpGetter // 节点和httpGetter映射
}

// NewHTTPPool 初始化服务端连接
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log 记录服务端日志
func (pool *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", pool.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 服务处理入口
func (pool *HTTPPool) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	// 1.如果URL不包含默认服务路径，抛出异常
	if !strings.HasPrefix(req.URL.Path, pool.basePath) {
		fmt.Println(pool.basePath)
		panic("HTTPPool serving unexpected path: " + req.URL.Path)
	}

	// 2.日志记录访问路径
	pool.Log("%s %s", req.Method, req.URL.Path)

	// 3.切割访问路径中的参数
	parts := strings.SplitN(req.URL.Path[len(pool.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(writer, "bad request", http.StatusBadRequest)
	}
	groupName := parts[0]
	key := parts[1]

	// 4.获取缓存分组获取缓存
	group := GetGroup(groupName)
	if group == nil {
		http.Error(writer, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Write(view.ByteSlice())
}

// Set 添加节点，方法实例化了一致性哈希算法，并且添加了传入的节点
func (pool *HTTPPool) Set(peers ...string) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	pool.peers = consistenthash.New(defaultReplicas, nil)
	pool.peers.Add(peers...)
	pool.httpGetters = make(map[string]*HttpGetter, len(peers))

	for _, peer := range peers {
		pool.httpGetters[peer] = &HttpGetter{baseURL: peer + pool.basePath}
	}
}

// PickPeer 实现缓存查询选择器，包装了一致性哈希算法的 Get() 方法，根据具体的 key，选择节点，返回节点对应的 HTTP 客户端。
func (pool *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	pool.mu.Lock()
	defer pool.mu.Unlock()
	peer := pool.peers.Get(key)
	if peer != "" && peer != pool.self {
		pool.Log("Pick peer %s", peer)
		return pool.httpGetters[peer], true
	}

	return nil, false
}

// HttpGetter 缓存远程获取器
type HttpGetter struct {
	baseURL string
}

// Get 远程获取缓存
func (h *HttpGetter) Get(group string, key string) ([]byte, error) {
	// 1.构建远程访问路径
	path := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	// 2.发起请求
	response, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	// 3.判断请求
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned:%v", response.StatusCode)
	}

	// 4.读取响应内容
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
