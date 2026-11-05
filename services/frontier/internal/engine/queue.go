package engine

import "sync"

// Item represents a URL candidate waiting to be crawled.
type Item struct {
	// JobID identifies the crawl job this URL belongs to.
	JobID string

	// URL is the normalized target address to crawl.
	URL string

	// Depth is the crawl distance from the seed URL.
	Depth int
}

// Queue is a thread-safe FIFO queue for crawl items.
type Queue struct {
	mu    sync.Mutex
	items []Item
}

// NewQueue creates an empty Queue.
func NewQueue() *Queue {
	return &Queue{}
}

// Push appends an item to the back of the queue.
func (q *Queue) Push(item Item) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.items = append(q.items, item)
}

// Pop removes and returns the front item, or false if the queue is empty.
func (q *Queue) Pop() (Item, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.items) == 0 {
		return Item{}, false
	}

	item := q.items[0]
	q.items[0] = Item{}
	q.items = q.items[1:]

	if len(q.items) == 0 {
		q.items = nil
	}

	return item, true
}

// Len returns the current number of items in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.items)
}
