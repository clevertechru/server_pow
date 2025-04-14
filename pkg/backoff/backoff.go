package backoff

import (
	"container/ring"
	"log"
	"net"
	"sync"
	"time"
)

type Queue struct {
	mu            sync.Mutex
	queue         *ring.Ring
	maxSize       int
	currentSize   int
	backoffBase   time.Duration
	maxBackoff    time.Duration
	retryAttempts map[string]int
}

type QueuedConn struct {
	conn      net.Conn
	attempts  int
	nextRetry time.Time
}

func NewQueue(maxSize int, baseDelay, maxDelay time.Duration) *Queue {
	return &Queue{
		maxSize:       maxSize,
		queue:         ring.New(maxSize),
		backoffBase:   baseDelay,
		maxBackoff:    maxDelay,
		retryAttempts: make(map[string]int),
	}
}

func (q *Queue) Add(conn net.Conn) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.currentSize >= q.maxSize {
		return false
	}

	addr := conn.RemoteAddr().String()
	attempts := q.retryAttempts[addr]
	delay := q.calculateBackoff(attempts)

	queued := &QueuedConn{
		conn:      conn,
		attempts:  attempts,
		nextRetry: time.Now().Add(delay),
	}

	q.queue.Value = queued
	q.queue = q.queue.Next()
	q.currentSize++
	q.retryAttempts[addr] = attempts + 1

	return true
}

func (q *Queue) Get() net.Conn {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.currentSize == 0 {
		return nil
	}

	start := q.queue
	for i := 0; i < q.currentSize; i++ {
		if queued, ok := q.queue.Value.(*QueuedConn); ok {
			if time.Now().After(queued.nextRetry) {
				q.queue.Value = nil
				q.currentSize--
				delete(q.retryAttempts, queued.conn.RemoteAddr().String())
				return queued.conn
			}
		}
		q.queue = q.queue.Next()
	}
	q.queue = start
	return nil
}

func (q *Queue) calculateBackoff(attempts int) time.Duration {
	if attempts == 0 {
		return 0
	}
	backoff := q.backoffBase * time.Duration(1<<uint(attempts-1))
	if backoff > q.maxBackoff {
		return q.maxBackoff
	}
	return backoff
}

func (q *Queue) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.currentSize
}

func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i := 0; i < q.currentSize; i++ {
		if queued, ok := q.queue.Value.(*QueuedConn); ok {
			if err := queued.conn.Close(); err != nil {
				log.Printf("Error closing queued connection: %v", err)
			}
		}
		q.queue.Value = nil
		q.queue = q.queue.Next()
	}
	q.currentSize = 0
	q.retryAttempts = make(map[string]int)
}
