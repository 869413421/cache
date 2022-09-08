package lru

import (
	"reflect"
	"testing"
)

type String string

func (str String) Len() int {
	return len(str)
}

func TestGet(t *testing.T) {
	lru := New(0, nil)
	lru.Add("key", String("1234"))
	value, ok := lru.Get("key")
	if !ok {
		t.Fatalf("get key fail")
	}

	if string(value.(String)) != "1234" {
		t.Fatalf("value not equal 1234")
	}

	value, ok = lru.Get("key2")
	if ok {
		t.Fatalf("get key error")
	}
}

func TestRemoveOldest(t *testing.T) {
	key1, key2, key3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cacheLen := len(key1 + v1 + key2 + v2)
	lru := New(int64(cacheLen), nil)
	lru.Add(key1, String(v1))
	lru.Add(key2, String(v2))
	lru.Add(key3, String(v3))

	if _, ok := lru.Get(key1); ok || lru.Len() != 2 {
		t.Fatalf("RemoveOldest Fail")
	}
}

func TestOnEvicted(t *testing.T) {
	var keys []string
	onEvicted := func(key string, value Value) {
		keys = append(keys, key)
	}
	key1, key2, key3, key4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := "value1", "value2", "value3", "value4"
	cacheLen := len(key1 + v1 + key2 + v2)
	lru := New(int64(cacheLen), onEvicted)
	lru.Add(key1, String(v1))
	lru.Add(key2, String(v2))
	lru.Add(key3, String(v3))
	lru.Add(key4, String(v4))

	expect := []string{"key1", "key2"}

	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s:%s", expect, keys)
	}
}
