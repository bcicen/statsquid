package agent

import (
	"sync"
)

//Streaming message queue
type StreamQ struct {
	queue [][]byte
	lock  sync.RWMutex
}

func newStreamQ() *StreamQ {
	return &StreamQ{
		queue: [][]byte{},
		lock:  sync.RWMutex{},
	}
}

func (q *StreamQ) add(line []byte) {
	q.lock.Lock()
	q.queue = append(q.queue, line)
	q.lock.Unlock()
}

func (q *StreamQ) isEmpty() bool {
	return (len(q.queue) < 1)
}

func (q *StreamQ) flush() [][]byte {
	q.lock.Lock()
	items := q.queue
	q.queue = [][]byte{}
	q.lock.Unlock()
	return items
}
