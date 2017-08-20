package server

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrJobTimedOut = errors.New("job request timed out")
)

type Task struct {
	args   []interface{}
	wg     sync.WaitGroup
	result interface{}
}

type Pool struct {
	concurrency int
	tasksChan   chan *Task
	wg          sync.WaitGroup
}

func (p *Pool) Size() int {
	return p.concurrency
}

func NewPool(concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
	}
}

func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker()
	}
}

func (p *Pool) Stop() {
	close(p.tasksChan)
	p.wg.Wait()
}

func (p *Pool) Add(args []interface{}) interface{} {
	t := Task{
		args: args,
		wg:   sync.WaitGroup{},
	}

	t.wg.Add(1)
	p.tasksChan <- &t

	return t
}

func (p *Pool) AddWithTimeout(args []interface{}, timeout time.Duration) (interface{}, error) {
	t := Task{
		args: args,
		wg:   sync.WaitGroup{},
	}

	t.wg.Add(1)
	select {
	case p.tasksChan <- &t:
		break
	case <-time.After(timeout):
		return nil, ErrJobTimedOut
	}

	return t, nil
}

func (p *Pool) runWorker() {
	for t := range p.tasksChan {
		t.result = t.work()
		t.wg.Done()
	}
	p.wg.Done()
}
