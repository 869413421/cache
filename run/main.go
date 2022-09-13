package main

import (
	"fmt"
	"github.com/869413421/cache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	cache.NewGroup("scores", 2<<10,
		cache.GetterFunc(
			func(key string) ([]byte, error) {
				log.Println("[SlowDb] search key", key)
				if v, ok := db[key]; ok {
					return []byte(v), nil
				}
				return nil, fmt.Errorf("%s not exist", key)
			},
		),
	)

	addr := ":9999"
	handler := cache.NewHTTPPool(addr)
	log.Println("cache is running at", addr)
	http.ListenAndServe(addr, handler)
}
