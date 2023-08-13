package redis

import (
	"container/list"
	"sync"
)

type LRUCache struct {
	sync.RWMutex
	Capacity int
	Keys     map[int]*list.Element
	Lists    *list.List
}

type lruPair struct {
	K, V int
}

func NewLRU(capacity int) LRUCache {
	return LRUCache{
		Capacity: capacity,
		Keys:     make(map[int]*list.Element),
		Lists:    list.New(),
	}
}

func (c *LRUCache) Get(key int) int {
	if val, ok := c.Keys[key]; ok {
		c.Lists.MoveToFront(val)
		return val.Value.(lruPair).V
	}
	return -1
}
func (c *LRUCache) Put(key, value int) {
	if val, ok := c.Keys[key]; ok {
		val.Value = lruPair{
			K: key,
			V: value,
		}
		c.Lists.MoveToFront(val)
		return
	}
	if c.Lists.Len() >= c.Capacity {
		delete(c.Keys, c.Lists.Back().Value.(lruPair).K)
		c.Lists.Remove(c.Lists.Back())
	}
	c.Keys[key] = c.Lists.PushFront(lruPair{key, value})
}
