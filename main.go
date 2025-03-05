package main

import (
	"fmt"
	"geektime-go-cache/cache"
)

func OnEvictedFunc(key string, val any) {
	fmt.Printf("onEvicted key: %s, val: %s\n", key, val)
}

func main() {
	b := cache.NewBuildInMemoryCache(cache.WithOnEvicted(OnEvictedFunc))
}
