package workerpool

import (
	"log"
	"net"
	"sync"
	"sync/atomic"

	"github.com/clevertechru/server_pow/pkg/metrics"
)

type Pool struct {
	workers       int
	tasks         chan net.Conn
	wg            sync.WaitGroup
	handler       func(net.Conn)
	isShutdown    bool
	mu            sync.RWMutex
	activeWorkers int32
}

func NewPool(workers int, handler func(net.Conn)) *Pool {
	if workers <= 0 {
		workers = 10 // default number of workers
	}

	p := &Pool{
		workers: workers,
		tasks:   make(chan net.Conn, workers), // buffer size = workers
		handler: handler,
	}

	// Update metrics
	metrics.WorkerPoolSize.Set(float64(workers))

	p.start()
	return p
}

func (p *Pool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool) worker() {
	atomic.AddInt32(&p.activeWorkers, 1)
	defer func() {
		atomic.AddInt32(&p.activeWorkers, -1)
		p.wg.Done()
	}()

	for conn := range p.tasks {
		if conn != nil {
			p.handler(conn)
		}
	}
}

func (p *Pool) Submit(conn net.Conn) bool {
	p.mu.RLock()
	if p.isShutdown {
		p.mu.RUnlock()
		return false
	}
	p.mu.RUnlock()

	// Try to submit without blocking
	select {
	case p.tasks <- conn:
		// Update queue size metric
		metrics.WorkerPoolQueueSize.Set(float64(len(p.tasks)))
		return true
	default:
		log.Printf("Worker pool is full, connection rejected")
		return false
	}
}

func (p *Pool) Shutdown() {
	p.mu.Lock()
	if !p.isShutdown {
		p.isShutdown = true
		close(p.tasks)
	}
	p.mu.Unlock()

	p.wg.Wait()
}
