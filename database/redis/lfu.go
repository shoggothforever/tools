package redis

import (
	"container/heap"
)

//
//type LFUCache struct {
//	Capacity int
//	Minimum  int
//	NodeMap  map[int]*list.Element
//	ListMap  map[int]*list.List
//}
//type lfuTuple struct {
//	K, V, F int
//}
//
//func Constructor(capacity int) LFUCache {
//	return LFUCache{
//		Capacity: capacity,
//		Minimum:  0,
//		NodeMap:  make(map[int]*list.Element),
//		ListMap:  make(map[int]*list.List),
//	}
//}

//func (c *LFUCache) Get(key int) int {
//	if val, ok := c.NodeMap[key]; !ok {
//		return -1
//	} else {
//		tuple := val.Value.(*lfuTuple)
//		c.ListMap[tuple.F].Remove(val)
//		tuple.F++
//		if _, ok := c.ListMap[tuple.F]; !ok {
//			c.ListMap[tuple.F] = list.New()
//		}
//		flist := c.ListMap[tuple.F]
//		c.NodeMap[key] = flist.PushFront(tuple)
//		if tuple.F-1 == c.Minimum && c.ListMap[tuple.F-1].Len() == 0 {
//			delete(c.ListMap, tuple.F-1)
//			c.Minimum += 1
//		}
//		return tuple.V
//	}
//}
//
//func (c *LFUCache) Put(key int, value int) {
//	//LFUCache非法
//	if c.Capacity == 0 {
//		return
//	}
//	//查询缓存中是否存在对应的key,若有则更新value值和frequency值
//	if val, ok := c.NodeMap[key]; ok {
//		tuple := val.Value.(*lfuTuple)
//		tuple.V = value
//		c.Get(key)
//		return
//	}
//	//如果不存在并且缓存满了,删除访问频率最低的链表的最后一个节点
//	if c.Capacity == len(c.NodeMap) {
//		curList := c.ListMap[c.Minimum]
//		backNode := curList.Back()
//		delete(c.NodeMap, backNode.Value.(*lfuTuple).K)
//		curList.Remove(backNode)
//	}
//	//插入节点
//	c.Minimum = 1
//	newTuple := &lfuTuple{
//		K: key,
//		V: value,
//		F: 1,
//	}
//	if _, ok := c.ListMap[1]; !ok {
//		c.ListMap[1] = list.New()
//	}
//	c.NodeMap[key] = c.ListMap[1].PushFront(newTuple)
//}

type PriorityQueue []*Item
type Item struct {
	K, V, F, TimeStamp, Index int
}

type LFUCache struct {
	Capacity      int
	TrueTime      int
	PriorityQueue PriorityQueue
	ItemMap       map[int]*Item
}

func Constructor(capacity int) LFUCache {
	return LFUCache{
		Capacity:      capacity,
		TrueTime:      0,
		PriorityQueue: PriorityQueue{},
		ItemMap:       make(map[int]*Item, capacity),
	}
}
func (q PriorityQueue) Len() int {
	return len(q)
}
func (q PriorityQueue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].Index = i
	q[j].Index = j
}
func (q PriorityQueue) Less(i, j int) bool {
	if q[i].F == q[j].F {
		return q[i].TimeStamp < q[j].TimeStamp
	}
	return q[i].F < q[j].F
}
func (q *PriorityQueue) Push(x interface{}) {
	n := len(*q)
	item := x.(*Item)
	item.Index = n
	*q = append(*q, item)
}

func (q *PriorityQueue) Pop() interface{} {
	old := *q
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*q = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (q *PriorityQueue) update(item *Item, value int, frequency int, count int) {
	item.V = value
	item.TimeStamp = count
	item.F = frequency
	heap.Fix(q, item.Index)
}
func (c *LFUCache) Get(key int) int {
	if c.Capacity == 0 {
		return -1
	}
	if val, ok := c.ItemMap[key]; ok {
		c.TrueTime++
		c.PriorityQueue.update(val, val.V, val.F+1, c.TrueTime)
		return val.V
	}
	return -1
}
func (c *LFUCache) Put(key, value int) {
	if c.Capacity == 0 {
		return
	}
	c.TrueTime++
	if val, ok := c.ItemMap[key]; ok {
		c.PriorityQueue.update(val, value, val.F+1, c.TrueTime)
		return
	}
	if len(c.PriorityQueue) == c.Capacity {
		item := heap.Pop(&c.PriorityQueue).(*Item)
		delete(c.ItemMap, item.K)
	}
	item := &Item{
		K:         key,
		V:         value,
		TimeStamp: c.TrueTime,
	}
	heap.Push(&c.PriorityQueue, item)
	c.ItemMap[key] = item
}
