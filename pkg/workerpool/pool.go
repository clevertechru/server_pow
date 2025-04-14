package workerpool

import (
	"log"
	"net"
	"sync"
)

type Pool struct {
	workers    int
	tasks      chan net.Conn
	wg         sync.WaitGroup
	handler    func(net.Conn)
	isShutdown bool
	mu         sync.RWMutex
}

func NewPool(workers int, handler func(net.Conn)) *Pool {
	if workers <= 0 {
		workers = 10 // default number of workers
	}

	p := &Pool{
		workers: workers,
		tasks:   make(chan net.Conn, workers*2), // buffer size = 2x workers
		handler: handler,
	}

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
	defer p.wg.Done()

	for conn := range p.tasks {
		if conn == nil {
			return
		}
		p.handler(conn)
	}
}

func (p *Pool) Submit(conn net.Conn) bool {
	p.mu.RLock()
	if p.isShutdown {
		p.mu.RUnlock()
		return false
	}
	p.mu.RUnlock()

	select {
	case p.tasks <- conn:
		return true
	default:
		log.Printf("Worker pool is full, connection rejected")
		return false
	}
}

func (p *Pool) Shutdown() {
	p.mu.Lock()
	p.isShutdown = true
	close(p.tasks)
	p.mu.Unlock()

	p.wg.Wait()
}
