package remotedialer

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
)

type backPressure struct {
	cond   sync.Cond
	c      *connection
	paused bool
	closed bool
}

func newBackPressure(c *connection) *backPressure {
	return &backPressure{
		cond: sync.Cond{
			L: &sync.Mutex{},
		},
		c:      c,
		paused: false,
	}
}

func (b *backPressure) OnPause() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.paused = true
	b.cond.Broadcast()
}

func (b *backPressure) Close() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.closed = true
	b.cond.Broadcast()
}

func (b *backPressure) OnResume() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	b.paused = false
	b.cond.Broadcast()
}

func (b *backPressure) Pause() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	if b.paused {
		return
	}
	b.c.Pause()
	b.paused = true
}

func (b *backPressure) Resume() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()
	if !b.paused {
		return
	}
	b.c.Resume()
	b.paused = false
}

func (b *backPressure) Wait() {
	b.cond.L.Lock()
	defer b.cond.L.Unlock()

	for !b.closed && b.paused {
		b.cond.Wait()
	}
}
