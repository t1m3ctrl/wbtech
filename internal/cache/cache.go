package cache

import (
	"container/heap"
	"encoding/json"
	"log/slog"
	"os"
	"sync"
	"time"
	"wbtech"
)

type OrderCache struct {
	capacity int
	mu       sync.RWMutex
	cache    map[string]*cacheItem
	filename string
	ttl      time.Duration
	pq       *priorityQueue
}

type cacheItem struct {
	key      string
	order    wbtech.Order
	lastUsed time.Time
	index    int
}

func NewOrderCache(capacity int, filename string, ttl time.Duration) *OrderCache {
	pq := make(priorityQueue, 0)
	oc := &OrderCache{
		capacity: capacity,
		cache:    make(map[string]*cacheItem),
		filename: filename,
		pq:       &pq,
		ttl:      ttl,
	}
	heap.Init(oc.pq)

	if filename != "" {
		oc.loadFromFile(filename)
	}

	go oc.cleanupExpiredItems()
	return oc
}

func (oc *OrderCache) Get(key string) (wbtech.Order, bool) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	if item, found := oc.cache[key]; found {
		item.lastUsed = time.Now()
		heap.Fix(oc.pq, item.index)
		return item.order, true
	}
	return wbtech.Order{}, false
}

func (oc *OrderCache) Set(order wbtech.Order) error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	key := order.OrderUID

	if len(oc.cache) >= oc.capacity {
		item := heap.Pop(oc.pq).(*cacheItem)
		delete(oc.cache, item.key)
	}

	item := &cacheItem{
		key:      key,
		order:    order,
		lastUsed: time.Now(),
	}
	oc.cache[key] = item
	heap.Push(oc.pq, item)

	return nil
}

func (oc *OrderCache) cleanupExpiredItems() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		oc.removeExpired(oc.ttl)
	}
}

func (oc *OrderCache) removeExpired(expiry time.Duration) {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	now := time.Now()
	for _, item := range oc.cache {
		if now.Sub(item.lastUsed) > expiry {
			heap.Remove(oc.pq, item.index)
			delete(oc.cache, item.key)
		}
	}
}

func (oc *OrderCache) loadFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close file", "filename", filename, "error", err)
		}
	}(file)

	var orders []wbtech.Order
	if err := json.NewDecoder(file).Decode(&orders); err != nil {
		return
	}

	for _, order := range orders {
		_ = oc.Set(order)
	}
}

func (oc *OrderCache) Close() error {
	oc.mu.Lock()
	defer oc.mu.Unlock()

	file, err := os.Create(oc.filename)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			slog.Error("failed to close file", "filename", oc.filename, "error", err)
		}
	}(file)

	var orders []wbtech.Order
	for _, item := range oc.cache {
		orders = append(orders, item.order)
	}

	return json.NewEncoder(file).Encode(orders)
}

// куча для ЛРУ
type priorityQueue []*cacheItem

func (pq *priorityQueue) Len() int {
	return len(*pq)
}

func (pq *priorityQueue) Less(i, j int) bool {
	return (*pq)[i].lastUsed.Before((*pq)[j].lastUsed)
}

func (pq *priorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*cacheItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
