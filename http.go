package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 默认访问前缀
const defaultBasePath = "/_cache/"

// HTTPPool 服务端连接
type HTTPPool struct {
	self     string
	basePath string
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
