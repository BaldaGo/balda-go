/**
 * @file pool.go
 * @brief pool of goroutined
 *
 * Contains Task and Pool types and methods to work with Pool
 */
package server

import (
	// System
	"sync"

	// Third-party

	// Project
	"github.com/BaldaGo/balda-go/logger"
)

/// Task intreface (see @class Type)
type TaskI interface {
	work() error ///< Goroutine
}

/**
 * @class Task
 * @brief Asyncronic task which works in goroutine
 *
 * It must have work method which start in goroutine
 * with arguments args and store it's result into result field
 */
type Task struct {
	args   []interface{}  ///< Argumets to call goroutine
	wg     sync.WaitGroup ///< Waight group
	result error          ///< Error if it occured
}

/**
 * @class Pool
 * @brief Pool of goroutines
 *
 * It contains concurrency goroutines and send new task into tasksChan,
 * when task is done, goroutine send result into resultsChan
 */
type Pool struct {
	concurrency int            ///< Number of goroutines in pool
	tasksChan   chan *Task     ///< Channel with tasks
	resultsChan chan *Task     ///< Channel with results
	wg          sync.WaitGroup ///< Waight group
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
func NewPool(concurrency int) *Pool {
	return &Pool{
		concurrency: concurrency,
		tasksChan:   make(chan *Task),
		resultsChan: make(chan *Task),
	}
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
	}
}

/**
 * @brief Waight while all workers stop
 * @return errs Array of error
 *
 * It close tasksChan and wait while all goroutines finished
 */
func (p *Pool) Stop() []error {
	close(p.tasksChan)
	close(p.resultsChan)
	p.wg.Wait()

	var errs []error
	for res := range p.resultsChan {
		if res.result != nil {
			errs = append(errs, res.result)
		}
	}
	return errs
}

/**
 * @brief Add new task into Pool
 * @param[in] args Arguments to call goroutine
 * @return t Added task
 */
func (p *Pool) Add(args []interface{}) TaskI {
	t := Task{
		args: args,
		wg:   sync.WaitGroup{},
	}

	t.wg.Add(1)
	p.tasksChan <- &t

	return t
}

/**
 * @brief Start worker
 */
func (p *Pool) runWorker() {
	for t := range p.tasksChan {
		t.result = t.work()
		p.resultsChan <- t
		t.wg.Done()
	}
	p.wg.Done()
}
