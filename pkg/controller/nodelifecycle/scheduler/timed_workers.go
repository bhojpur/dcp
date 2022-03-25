package scheduler

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// WorkArgs keeps arguments that will be passed to the function executed by the worker.
type WorkArgs struct {
	NamespacedName types.NamespacedName
}

// KeyFromWorkArgs creates a key for the given `WorkArgs`
func (w *WorkArgs) KeyFromWorkArgs() string {
	return w.NamespacedName.String()
}

// NewWorkArgs is a helper function to create new `WorkArgs`
func NewWorkArgs(name, namespace string) *WorkArgs {
	return &WorkArgs{types.NamespacedName{Namespace: namespace, Name: name}}
}

// TimedWorker is a responsible for executing a function no earlier than at FireAt time.
type TimedWorker struct {
	WorkItem  *WorkArgs
	CreatedAt time.Time
	FireAt    time.Time
	Timer     *time.Timer
}

// CreateWorker creates a TimedWorker that will execute `f` not earlier than `fireAt`.
func CreateWorker(args *WorkArgs, createdAt time.Time, fireAt time.Time, f func(args *WorkArgs) error) *TimedWorker {
	delay := fireAt.Sub(createdAt)
	if delay <= 0 {
		go f(args)
		return nil
	}
	timer := time.AfterFunc(delay, func() { f(args) })
	return &TimedWorker{
		WorkItem:  args,
		CreatedAt: createdAt,
		FireAt:    fireAt,
		Timer:     timer,
	}
}

// Cancel cancels the execution of function by the `TimedWorker`
func (w *TimedWorker) Cancel() {
	if w != nil {
		w.Timer.Stop()
	}
}

// TimedWorkerQueue keeps a set of TimedWorkers that are still wait for execution.
type TimedWorkerQueue struct {
	sync.Mutex
	// map of workers keyed by string returned by 'KeyFromWorkArgs' from the given worker.
	workers  map[string]*TimedWorker
	workFunc func(args *WorkArgs) error
}

// CreateWorkerQueue creates a new TimedWorkerQueue for workers that will execute
// given function `f`.
func CreateWorkerQueue(f func(args *WorkArgs) error) *TimedWorkerQueue {
	return &TimedWorkerQueue{
		workers:  make(map[string]*TimedWorker),
		workFunc: f,
	}
}

func (q *TimedWorkerQueue) getWrappedWorkerFunc(key string) func(args *WorkArgs) error {
	return func(args *WorkArgs) error {
		err := q.workFunc(args)
		q.Lock()
		defer q.Unlock()
		if err == nil {
			// To avoid duplicated calls we keep the key in the queue, to prevent
			// subsequent additions.
			q.workers[key] = nil
		} else {
			delete(q.workers, key)
		}
		return err
	}
}

// AddWork adds a work to the WorkerQueue which will be executed not earlier than `fireAt`.
func (q *TimedWorkerQueue) AddWork(args *WorkArgs, createdAt time.Time, fireAt time.Time) {
	key := args.KeyFromWorkArgs()
	klog.V(4).Infof("Adding TimedWorkerQueue item %v at %v to be fired at %v", key, createdAt, fireAt)

	q.Lock()
	defer q.Unlock()
	if _, exists := q.workers[key]; exists {
		klog.Warningf("Trying to add already existing work for %+v. Skipping.", args)
		return
	}
	worker := CreateWorker(args, createdAt, fireAt, q.getWrappedWorkerFunc(key))
	q.workers[key] = worker
}

// CancelWork removes scheduled function execution from the queue. Returns true if work was cancelled.
func (q *TimedWorkerQueue) CancelWork(key string) bool {
	q.Lock()
	defer q.Unlock()
	worker, found := q.workers[key]
	result := false
	if found {
		klog.V(4).Infof("Cancelling TimedWorkerQueue item %v at %v", key, time.Now())
		if worker != nil {
			result = true
			worker.Cancel()
		}
		delete(q.workers, key)
	}
	return result
}

// GetWorkerUnsafe returns a TimedWorker corresponding to the given key.
// Unsafe method - workers have attached goroutines which can fire after this function is called.
func (q *TimedWorkerQueue) GetWorkerUnsafe(key string) *TimedWorker {
	q.Lock()
	defer q.Unlock()
	return q.workers[key]
}
