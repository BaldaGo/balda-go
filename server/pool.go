/**
 * @file pool.go
 * @brief pool of goroutined
 *
 * Contains Task and Pool types and methods to work with Pool
 */
package server

import (
	// System
	"context"
	"sync"
	"time"
	// Third-party
	// Project
	"github.com/BaldaGo/balda-go/logger"
)

/// Task intreface (see @class Type)
type Func func(context.Context) error ///< Goroutine
type key int

const (
	ConnKey key = 0 + iota
	ServerKey
	ChanKey
)

/**
 * @class Task
 * @brief Asyncronic task which works in goroutine
 *
 * It must have Work method which start in goroutine
 * with arguments args and store it's result into result field
 */
type Task struct {
	work   Func            ///< Callback
	ctx    context.Context ///< Context of calling goroutine
	result error           ///< Error if it occured
}

/**
 * @class Pool
 * @brief Pool of goroutines
 *
 * It contains concurrency goroutines and send new task into tasksChan,
 * when task is done, goroutine send result into resultsChan
 */
type Pool struct {
	concurrency int                  ///< Number of goroutines in pool
	tasksChan   chan *Task           ///< Channel with tasks
	resultsChan chan error           ///< Channel with results
	wg          sync.WaitGroup       ///< Waight group
	cancels     []context.CancelFunc ///< Array of functions stops working
}

func SyncFuncWithTimeout(f Func, ctx context.Context, timeout time.Duration) error {
	ch := make(chan error)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	go func(chan error) { ch <- f(ctx) }(ch)
	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

/**
 * @brief Return size of Pool
 * @return Size of Pool
 */
func (p *Pool) Size() int {
	return p.concurrency
}

/**
 * @brief Create new Pool object
 * @param[in] concurrency Number of goroutines in pool
 * @return Pointer to created Pool
 */
func NewPool(concurrency int) Pool {
	pool := Pool{
		concurrency: concurrency,
		tasksChan:   make(chan *Task, concurrency),
		resultsChan: make(chan error, concurrency),
		cancels:     make([]context.CancelFunc, 0, concurrency),
	}

	return pool
}

/**
 * @brief Start this Pool
 *
 * Runs all workers
 */
func (p *Pool) Run() {
	for i := 0; i < p.concurrency; i++ {
		p.wg.Add(1)
		go p.runWorker()
		go p.collectResultsAndLog()
	}
}

/**
 * @brief Waight while all workers stop
 * @return errs Array of error
 *
 * It close tasksChan and wait while all goroutines finished
 */
func (p *Pool) Stop() {
	close(p.tasksChan)
	for _, i := range p.cancels {
		i()
	}
	p.wg.Wait()
	close(p.resultsChan)
}

/**
 * @brief Add new task into Pool
 * @param[in] args Arguments to call goroutine
 * @return t Added task
 */
func (p *Pool) Add(work Func, ctx context.Context) {
	t := &Task{
		work: work,
		ctx:  ctx,
	}

	p.tasksChan <- t
}

func (p *Pool) collectResultsAndLog() {
	for err := range p.resultsChan {
		if err != nil {
			logger.Log.Critical("Work failed: ", err.Error())
		}
	}
}

/**
 * @brief Start worker
 */
func (p *Pool) runWorker() {
	defer p.wg.Done()
	for t := range p.tasksChan {
		p.resultsChan <- t.work(t.ctx)
	}
}
