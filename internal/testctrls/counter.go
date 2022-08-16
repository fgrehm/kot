package testctrls

import (
	"sync"
)

type SafeCounter interface {
	Value() int
	Increment()
	Decrement()
	Reset()
}

type atomicCounter struct {
	mu    sync.Mutex
	value int
}

var Counter SafeCounter = &atomicCounter{}

func (c *atomicCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *atomicCounter) Decrement() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value--
}

func (c *atomicCounter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

func (c *atomicCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
