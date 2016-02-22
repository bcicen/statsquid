package agent

import (
	"sync"

	"github.com/vektorlab/statsquid/models"
)

//Streaming message queue
type StatQ struct {
	queue []*models.StatSquidStat
	lock  sync.RWMutex
}

func newStatQ() *StatQ {
	return &StatQ{
		queue: []*models.StatSquidStat{},
		lock:  sync.RWMutex{},
	}
}

func (q *StatQ) add(s *models.StatSquidStat) {
	q.lock.Lock()
	q.queue = append(q.queue, s)
	q.lock.Unlock()
}

func (q *StatQ) isEmpty() bool {
	return (len(q.queue) < 1)
}

func (q *StatQ) flush() []byte {
	q.lock.Lock()
	packed := models.PackStats(q.queue)
	q.queue = nil
	q.lock.Unlock()
	return packed
}
