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
	"sync/atomic"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	testVal := int32(0)
	wg := sync.WaitGroup{}
	wg.Add(5)
	queue := CreateWorkerQueue(func(args *WorkArgs) error {
		atomic.AddInt32(&testVal, 1)
		wg.Done()
		return nil
	})
	now := time.Now()
	queue.AddWork(NewWorkArgs("1", "1"), now, now)
	queue.AddWork(NewWorkArgs("2", "2"), now, now)
	queue.AddWork(NewWorkArgs("3", "3"), now, now)
	queue.AddWork(NewWorkArgs("4", "4"), now, now)
	queue.AddWork(NewWorkArgs("5", "5"), now, now)
	// Adding the same thing second time should be no-op
	queue.AddWork(NewWorkArgs("1", "1"), now, now)
	queue.AddWork(NewWorkArgs("2", "2"), now, now)
	queue.AddWork(NewWorkArgs("3", "3"), now, now)
	queue.AddWork(NewWorkArgs("4", "4"), now, now)
	queue.AddWork(NewWorkArgs("5", "5"), now, now)
	wg.Wait()
	lastVal := atomic.LoadInt32(&testVal)
	if lastVal != 5 {
		t.Errorf("Expected testVal = 5, got %v", lastVal)
	}
}

func TestExecuteDelayed(t *testing.T) {
	testVal := int32(0)
	wg := sync.WaitGroup{}
	wg.Add(5)
	queue := CreateWorkerQueue(func(args *WorkArgs) error {
		atomic.AddInt32(&testVal, 1)
		wg.Done()
		return nil
	})
	now := time.Now()
	then := now.Add(10 * time.Second)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	wg.Wait()
	lastVal := atomic.LoadInt32(&testVal)
	if lastVal != 5 {
		t.Errorf("Expected testVal = 5, got %v", lastVal)
	}
}

func TestCancel(t *testing.T) {
	testVal := int32(0)
	wg := sync.WaitGroup{}
	wg.Add(3)
	queue := CreateWorkerQueue(func(args *WorkArgs) error {
		atomic.AddInt32(&testVal, 1)
		wg.Done()
		return nil
	})
	now := time.Now()
	then := now.Add(10 * time.Second)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	queue.CancelWork(NewWorkArgs("2", "2").KeyFromWorkArgs())
	queue.CancelWork(NewWorkArgs("4", "4").KeyFromWorkArgs())
	wg.Wait()
	lastVal := atomic.LoadInt32(&testVal)
	if lastVal != 3 {
		t.Errorf("Expected testVal = 3, got %v", lastVal)
	}
}

func TestCancelAndReadd(t *testing.T) {
	testVal := int32(0)
	wg := sync.WaitGroup{}
	wg.Add(4)
	queue := CreateWorkerQueue(func(args *WorkArgs) error {
		atomic.AddInt32(&testVal, 1)
		wg.Done()
		return nil
	})
	now := time.Now()
	then := now.Add(10 * time.Second)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	queue.AddWork(NewWorkArgs("1", "1"), now, then)
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	queue.AddWork(NewWorkArgs("3", "3"), now, then)
	queue.AddWork(NewWorkArgs("4", "4"), now, then)
	queue.AddWork(NewWorkArgs("5", "5"), now, then)
	queue.CancelWork(NewWorkArgs("2", "2").KeyFromWorkArgs())
	queue.CancelWork(NewWorkArgs("4", "4").KeyFromWorkArgs())
	queue.AddWork(NewWorkArgs("2", "2"), now, then)
	wg.Wait()
	lastVal := atomic.LoadInt32(&testVal)
	if lastVal != 4 {
		t.Errorf("Expected testVal = 4, got %v", lastVal)
	}
}
